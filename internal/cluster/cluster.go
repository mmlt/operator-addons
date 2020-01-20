// Access target k8s cluster.
package cluster

import (
	"crypto/tls"
	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"
	"github.com/mmlt/operator-addons/internal/exe"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Cluster represents a k8s cluster to manage.
type Cluster struct {
	// name is the name of the target cluster.
	// Typically this is also the metadata.name of the Cluster CR.
	Name string
	// Server is the url of the k8s API server.
	Server string
	// tempDir is the $PWD and $HOME when executing commands against the target cluster.
	// tempDir includes an .kube/config files.
	Path string
	// LastUpdate the cluster is reconciled.
	LastUpdate time.Time

	// Client is used to communicate with the target cluster.
	client *kubernetes.Clientset
	// Log is cluster specific logger.
	log logr.Logger
}

func New(name string, log logr.Logger) (*Cluster, error) {
	c := &Cluster{
		Name: name,
		Path: filepath.Join(os.TempDir(), name),
		log:  log.WithName("Cluster"),
	}

	// create .kube/config directory
	p := filepath.Join(c.Path, ".kube")
	err := os.MkdirAll(p, 0755)
	if err != nil {
		return nil, err
	}
	log.V(2).Info("Create dir", "path", p)

	return c, nil
}

// Ping checks if server is responding.
func (c *Cluster) Ping() bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout: time.Second,
		}).DialContext,
	}
	client := &http.Client{Transport: tr}
	req, err := http.NewRequest("GET", c.Server, nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return true
}

// SetServerCoordinates set all parameters to access an target API Server.
// Must provide password or clientCert+clientKey.
func (c *Cluster) SetServerCoordinates(url string, serverCA []byte, user, password string, clientCert, clientKey []byte) error {
	c.Server = url

	// Create kube config
	u := &api.AuthInfo{}
	if password != "" {
		u.Username = user
		u.Password = password
	} else {
		u.ClientCertificateData = clientCert
		u.ClientKeyData = clientKey
	}

	kc := api.Config{
		Kind:        "Config",
		APIVersion:  "v1",
		Preferences: api.Preferences{},
		Clusters: map[string]*api.Cluster{
			c.Name: {
				Server:                   c.Server,
				CertificateAuthorityData: serverCA,
			},
		},
		AuthInfos: map[string]*api.AuthInfo{
			user: u,
		},
		Contexts: map[string]*api.Context{
			"default": &api.Context{
				Cluster:  c.Name,
				AuthInfo: user,
			},
		},
		CurrentContext: "default",
	}

	d, err := clientcmd.Write(kc)
	if err != nil {
		return err
	}

	p := filepath.Join(c.Path, ".kube", "config")
	err = ioutil.WriteFile(p, d, 0755)
	if err != nil {
		return err
	}
	c.log.V(2).Info("Write file", "path", p)

	// Create clientset from kube/config
	config, err := clientcmd.BuildConfigFromFlags("", p)
	if err != nil {
		return err
	}
	c.log.V(2).Info("Read config", "path", p)
	// create the clientset
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	c.log.V(3).Info("Created client")

	c.client = client

	return nil
}

const CMName = "clusterops-state"
const CMNamespace = "kube-system"

// GetState reads state from target cluster.
func (c *Cluster) GetState() (map[string]string, error) {
	cm, err := c.client.CoreV1().ConfigMaps(CMNamespace).Get(CMName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	c.log.V(2).Info("GetState", "state", cm.Data)
	return cm.Data, nil
}

// PutState writes state to target cluster.
func (c *Cluster) PutState(data map[string]string) error {
	c.log.V(2).Info("PutState", "state", data)

	// check for existing ConfigMap
	var create bool
	cm, err := c.client.CoreV1().ConfigMaps(CMNamespace).Get(CMName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		create = true
	} else if err != nil {
		return err
	}

	if create {
		// create new configmap
		_, err = c.client.CoreV1().ConfigMaps(CMNamespace).Create(&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      CMName,
				Namespace: CMNamespace,
			},
			Data: data,
		})
		if err != nil {
			return err
		}
	} else {
		// update existing configmap
		cm.Data = data
		_, err = c.client.CoreV1().ConfigMaps(CMNamespace).Update(cm)
		if err != nil {
			return err
		}
	}

	return nil
}

// RunShell runs cmd in a shell with optional values and extra environment variables.
//
// Values is an arbiratry structure that is provided as values.yaml and as environment variables
// in the form VALUE_PATH_TO_KEY.
//
// ExtraEnv is a list of "KEY=Value".
//
// The environment contains the config needed to run kubectl against the target cluster.
// Environment variables:
//	HOME		path to home directory (PWD = HOME)
//	+extraEnv
//  +values flattened add changed to uppercase.
//
// File system:
// 	$HOME/
//		.kube/config
//		values.yaml
//
func (c *Cluster) RunShell(cmd string, values interface{}, extraEnv []string) error {
	err := c.writeValuesYaml(values)
	if err != nil {
		return err
	}

	// Collect environment variables
	env := append(os.Environ(), MapToEnv(values, "VALUE_")...)
	env = append(env, extraEnv...)
	env = append(env, "HOME="+c.Path)

	opt := exe.Opt{
		Dir: c.Path,
		Env: env,
	}

	_, _, err = exe.Run("bash", exe.Args{"-c", cmd}, opt, c.log)
	if err != nil {
		return err
	}

	return nil
}

// WriteValuesYaml write a values.yaml file with 'data' in $HOME
func (c *Cluster) writeValuesYaml(data interface{}) error {
	d, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	p := filepath.Join(c.Path, "values.yaml")
	err = ioutil.WriteFile(p, d, 0755)
	if err != nil {
		return err
	}
	c.log.V(2).Info("Write file", "path", p)

	return nil
}
