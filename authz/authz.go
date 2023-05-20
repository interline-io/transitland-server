package authz

type AuthzConfig struct {
	Auth0Domain       string
	Auth0ClientID     string
	Auth0ClientSecret string
	FGAStoreID        string
	FGAModelID        string
	FGAEndpoint       string
	FGATestModelPath  string
	FGATestTuplesPath string
	GlobalAdmin       string
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
