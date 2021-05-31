package auth

// will be init at compile time with the -X parameter

// configuration
var (
	JwtSecret string

	GithubClientID     string
	GithubClientSecret string
	GithubCallbackURL  string

	FacebookClientID     string
	FacebookClientSecret string
	FacebookCallbackURL  string

	JWTSecret string
)
