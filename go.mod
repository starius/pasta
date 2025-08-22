module github.com/starius/pasta

go 1.23.0

toolchain go1.24.0

require (
	github.com/golang/snappy v1.0.0
	github.com/monperrus/crawler-user-agents v1.18.0
	github.com/robfig/humanize v0.0.0-20130801072920-4123e5c9f2f9
	github.com/tyler-smith/go-bip39 v1.1.0
	gitlab.com/NebulousLabs/entropy-mnemonics v0.0.0-20181018051301-7532f67e3500
	gitlab.com/NebulousLabs/fastrand v0.0.0-20181126182046-603482d69e40
	gitlab.com/starius/deallocate v0.0.0-20190713141632-605b24537969
	gitlab.com/starius/encrypt-autocert-cache v0.1.2
	gitlab.com/starius/fpe v0.0.0-20181110234326-b113c8214a5f
	golang.org/x/crypto v0.41.0
	golang.org/x/net v0.43.0
	golang.org/x/term v0.34.0
	google.golang.org/protobuf v1.36.8
)

require (
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

// tyler-smith's repo was removed. Use a fork.
replace github.com/tyler-smith/go-bip39 => github.com/alexvec/go-bip39 v1.1.0
