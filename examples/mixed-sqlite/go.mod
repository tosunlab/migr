module example.com/mixed-sqlite

go 1.24

require github.com/tosunlab/migrasyon/sdk/go v0.0.0

// Inside the migrasyon repository the SDK is this checkout's. In a project of your own, delete this line
// and run `go get github.com/tosunlab/migrasyon/sdk/go`.
replace github.com/tosunlab/migrasyon/sdk/go => ../../sdk/go
