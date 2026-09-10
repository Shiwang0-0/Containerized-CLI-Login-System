### Run SQL using docker compose

```bash
docker compose up

# migrate
migrate   -path ./migrations  -database "mysql://cli_user:cli_password@tcp(localhost:3307)/cli_db" up

docker compose down # volume still persist
docker compose down -v # volume deleted

# to view the running container
docker compose ps

# to run commands inside the container
docker compose exec mysql mysql -u <user> -p<password> <db_name>
```
Note: The migration command is present in the docker compose file itself, so write the migration into a new incremental file so that the previous data is still present in the volume and you dont have to delete the volume.


# Some good practices that I followed

### Repository Pattern  
Whole codebase is divided keeping the repository pattern in mind, this allows seperation of concerns for different usecases.

### Retrying if Database connection failed  
```go
	for i := 1; i <= 5; i++ {
		if err := db.Ping(); err != nil {
			log.Printf("MySQL connection attempt %d/5 failed: %v", i, err)
			time.Sleep(2 * time.Second)
		} else {
			log.Println("MySQL connected")
			connected = true
			break
		}
	}
```

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