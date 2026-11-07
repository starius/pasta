module github.com/starius/pasta

go 1.27.3

require (
	github.com/golang/snappy v1.0.0
	github.com/monperrus/crawler-user-agents v1.66.0
	github.com/robfig/humanize v0.0.0-20130801072920-4123e5c9f2f9
	github.com/tyler-smith/go-bip39 v1.1.0
	gitlab.com/NebulousLabs/entropy-mnemonics v0.0.0-20181018051301-7532f67e3500
	gitlab.com/starius/deallocate v0.0.0-20190713141632-605b24537969
	gitlab.com/starius/encrypt-autocert-cache v0.1.2
	gitlab.com/starius/fpe v0.0.0-20181110234326-b113c8214a5f
	golang.org/x/crypto v0.58.1-0.20261009181718-2f1f834330b2 // Fix for GO-2026-5932; newer than the latest tagged release.
	golang.org/x/net v0.62.0
	golang.org/x/term v0.47.0
	google.golang.org/protobuf v1.36.12
)

require (
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

// tyler-smith's repo was removed. Use a fork.
replace github.com/tyler-smith/go-bip39 => github.com/alexvec/go-bip39 v1.1.0
