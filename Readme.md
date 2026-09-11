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


### TOTP 2FA Implementation
After successful login, user will have a choice to `enable-2fa`.
User will get 2 choice
1. Scan the QR using the authenticator app  
2. Enter the App name and key manually (display on the terminal)  

Once 2FA is enabled, every time user tries to login, the 6 digit authentication code will be required to successfully login.  
User can disable the 2FA using `disable-2fa` command, but only if the user is already logged in and knows the password.  


### Account Lockout Implementation  
There are two counters for failed attempts. `FailedAttempts` for wrong password guess and `TOTPFailedAttempts` for wrong 2FA code guess. 
There is a base number of attemps that user can do (currently set to 5), after this will be an exponential increase in time as the number of failed attempt increases.  

```go
	// the last lockout happened before 24hr, that means the level of the waiting should increase
	// otherwise reset the level to 0
	/*
		base duration = x min
		First lockout                -----> level 0 ----->   x*(2^0) => x min lock duration
		Second lockout within 24hr   -----> level 1 -----> 	 x*(2^1) => 2x min lock duration
		Third lockout within 24hr    -----> level 2 ----->   x*(2^2) => 4x min lock duration
		.
		.
		.
		Nth lockout within 24hr      -----> level N-1 -----> min(24 hr, x*(2^(N-1))) => 24 hr lock duration
		.
		24 hr passed ===> level reset to 0
	*/
```

This ensures that the account has a graceful increase in duration of lockout. We dont want immediate shut down of the account. similary we dont want infinte lock duration so there is a cap of 24hr after the last lockout after which the exponential level of the waiting resets.  
Once the user successfully logs in (after TOTP if any) then the failure counters resets to 0.

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

### Gracefully increasing the duration of lockout  
User account doesnt automatically gets lockout, the duration of lockout gracefully increases until it reaches its maximum limit of 24 hr within the last lockout.  

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