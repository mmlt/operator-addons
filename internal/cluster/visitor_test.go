package cluster

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
	"strings"
	"testing"
)

func Example_visit_a_map() {
	s := `
value1: v_one
elem1:
  elmen2: e_two
`
	v := make(map[interface{}]interface{})
	_ = yaml.NewDecoder(strings.NewReader(s)).Decode(v)

	Visit(v, func(p []string, v string) {
		fmt.Println(p, v)
	})
	// Output:
	// [value1] v_one
	// [elem1 elmen2] e_two
}

func Example_visit_a_slice() {
	s := `
elem1:
- e_one
- e_two
`
	v := make(map[interface{}]interface{})
	_ = yaml.NewDecoder(strings.NewReader(s)).Decode(v)

	Visit(v, func(p []string, v string) {
		fmt.Println(p, v)
	})
	// Output:
	// [elem1 0] e_one
	// [elem1 1] e_two
}

func Example_yaml_to_environment_variables() {
	s := `
value1: v_one
elem1:
- e_one
- e_two
`
	v := make(map[interface{}]interface{})
	_ = yaml.NewDecoder(strings.NewReader(s)).Decode(v)
	fmt.Println(
		MapToEnv(v, "prefix_"))
	// Unordered output:
	// [PREFIX_VALUE1=v_one PREFIX_ELEM1_0=e_one PREFIX_ELEM1_1=e_two]
}

// Example_visitfn_env is an example of a VisitFn to create environment variables.
func Example_visit_function() {
	var result string
	fn := func(p []string, v string) {
		result = fmt.Sprintf("%s=%s", strings.ToUpper(strings.Join(p, "_")), v)
	}
	fn([]string{"p1", "p2"}, "value")
	fmt.Println(result)
	// Output:
	// P1_P2=value
}

func TestVisitMap(t *testing.T) {
	s := `
value1: v_one
elem1:
  elmen2: e_two
`
	//var v interface{}
	v := make(map[interface{}]interface{})
	err := yaml.NewDecoder(strings.NewReader(s)).Decode(v)
	assert.NoError(t, err)

	Visit(v, func(p []string, v string) {
		fmt.Println(p, v)
	})
}
