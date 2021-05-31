module github.com/InjectiveLabs/peggo

go 1.16

require (
	github.com/InjectiveLabs/etherman v1.7.0
	github.com/InjectiveLabs/sdk-go v1.20.12
	github.com/alexcesaro/statsd v2.0.0+incompatible
	github.com/avast/retry-go v3.0.0+incompatible
	github.com/cosmos/cosmos-sdk v0.42.5
	github.com/ethereum/go-ethereum v1.10.2
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/jawher/mow.cli v1.2.0
	github.com/onsi/ginkgo v1.15.1
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v1.2.0
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.10
	github.com/xlab/closer v0.0.0-20190328110542-03326addb7c2
	github.com/xlab/suplog v1.3.0
	golang.org/x/crypto v0.0.0-20201221181555-eec23a3978ad
	google.golang.org/genproto v0.0.0-20210114201628-6edceaf6022f
	google.golang.org/grpc v1.37.1
	gopkg.in/alexcesaro/statsd.v2 v2.0.0 // indirect
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
