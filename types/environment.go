package types

// Environment The server environment, either sandbox or production.
type Environment string

const (
	EnvironmentSandbox    Environment = "Sandbox"    // Indicates that the notification applies to testing in the sandbox environment.
	EnvironmentProduction Environment = "Production" // Indicates that the notification applies to the production environment.
)

func (e Environment) String() string {
	return string(e)
}
