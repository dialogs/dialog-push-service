module github.com/dialogs/dialog-push-service

go 1.13

// fix hash error
replace software.sslmate.com/src/go-pkcs12 => software.sslmate.com/src/go-pkcs12 v0.0.0-20190322163127-6e380ad96778

require (
	github.com/dialogs/dialog-go-lib v1.1.14
	github.com/gogo/protobuf v1.3.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/mailru/easyjson v0.7.0
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.0
	github.com/sideshow/apns2 v0.0.0-20171218084920-df275e5c35d2
	github.com/spf13/viper v1.4.0
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20190613194153-d28f0bde5980
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/grpc v1.24.0
)
