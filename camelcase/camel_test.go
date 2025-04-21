package camelcase

import (
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestToUpperCamel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Empty string", "", ""},
		{"Single word", "hello", "Hello"},
		{"Snake case", "hello_world", "HelloWorld"},
		{"lower camel case", "helloWorld", "HelloWorld"},
		{"upper camel case", "HelloWorld", "HelloWorld"},
		{"upper camel case", "HelloWorld_1", "Helloworld1"}, // all parts split by underscore should be recognized as a word
		{"All caps", "HTTP_SERVER", "HTTPServer"},
		{"Mixed case", "mySQL_Query", "MySQLQuery"},
		{"With numbers", "user_id_2", "UserID2"},
		{"Abbreviations", "http_request", "HTTPRequest"},
		{"Unicode", "こんにちは_世界", "こんにちは世界"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ToUpperCamel(tt.input)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestToLowerCamel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Empty string", "", ""},
		{"Single word", "Hello", "hello"},
		{"Snake case", "hello_world", "helloWorld"},
		{"All caps", "HTTP_SERVER", "httpServer"},
		{"With numbers", "USER_ID_2", "userID2"},
		{"Abbreviations", "HTTP_REQUEST", "httpRequest"},
		{"Mixed case", "MySQL_Query", "mysqlQuery"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ToLowerCamel(tt.input)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestToUnderScore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"Empty string", "", ""},
		{"Single word", "Hello", "hello"},
		{"Camel case", "helloWorld", "hello_world"},
		{"All caps", "HttpServer", "http_server"},
		{"All caps", "HTTPServer", "http_server"},
		{"With numbers", "UserID2", "user_id_2"},
		{"Abbreviations", "HTTPRequest", "http_request"},
		{"Mixed case", "MySQLQuery", "mysql_query"},
		{"Consecutive caps", "MySSHKey", "my_ssh_key"},
		{"Unicode", "こんにちはWorld", "こんにちは_world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ToUnderScore(tt.input)
			assert.Equal(t, got, tt.want)
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("Multiple underscores", func(t *testing.T) {
		t.Parallel()

		got := ToUpperCamel("hello__world")
		assert.Equal(t, got, "HelloWorld")
	})

	t.Run("Leading underscore", func(t *testing.T) {
		t.Parallel()

		got := ToUpperCamel("_hello_world")
		assert.Equal(t, got, "HelloWorld")
	})

	t.Run("All letters uppercase", func(t *testing.T) {
		t.Parallel()

		got := ToUnderScore("HELLOWORLD")
		assert.Equal(t, got, "helloworld")
	})
}

func BenchmarkToUpperCamel(b *testing.B) {
	testString := "hello_world_this_is_a_benchmark_test"
	for i := 0; i < b.N; i++ {
		ToUpperCamel(testString)
	}
}

func BenchmarkToLowerCamel(b *testing.B) {
	testString := "HELLO_WORLD_THIS_IS_A_BENCHMARK_TEST"
	for i := 0; i < b.N; i++ {
		ToLowerCamel(testString)
	}
}

func BenchmarkToUnderScore(b *testing.B) {
	testString := "HelloWorldThisIsABenchmarkTest"
	for i := 0; i < b.N; i++ {
		ToUnderScore(testString)
	}
}

func FuzzCamelCase(f *testing.F) {
	f.Add("hello_world")
	f.Add("HTTPRequest")
	f.Add("userID2")

	f.Fuzz(func(t *testing.T, s string) {
		if !utf8.ValidString(s) {
			t.Skip()
		}

		// Test roundtrip conversions
		upper := ToUpperCamel(s)
		roundtrip := ToUnderScore(upper)
		upper2 := ToUpperCamel(roundtrip)
		assert.Equal(t, upper, upper2)

		// Check all letters are valid
		for _, r := range upper {
			assert.True(t, unicode.IsLetter(r) || unicode.IsDigit(r))
		}
	})
}
