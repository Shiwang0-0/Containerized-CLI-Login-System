### Validator for Username and Password input

using `go-playground/validator/v10` library for input validation and created a custom regex validator for username

```go
const (
	usernameRules = "required,min=3,max=32,username"
	passwordRules = "required,min=12,max=72"
)

var (
	validate      = validator.New()
	usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
)

func init() {
    // custom regex validator for username
	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})
}
```