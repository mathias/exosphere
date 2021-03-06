package types_test

import (
	"github.com/Originate/exosphere/src/types"
	"github.com/Originate/exosphere/test/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AppConfig", func() {
	var appConfig types.AppConfig

	var _ = Describe("ValidateFields", func() {
		It("should throw an error when AppConfig is missing fields in production", func() {
			appConfig = types.AppConfig{
				Remote: types.AppRemoteConfig{
					URL:       "originate.com",
					AccountID: "123",
					Region:    "us-west-2",
				},
			}
			err := appConfig.Remote.ValidateFields()
			Expect(err).To(HaveOccurred())
			expectedErrorString := "application.yml missing required field 'production.SslCertificateArn'"
			Expect(err.Error()).To(ContainSubstring(expectedErrorString))
		})

		It("should not throw an error when AppConfig isn't missing fields", func() {
			appConfig = types.AppConfig{
				Remote: types.AppRemoteConfig{
					URL:               "originate.com",
					AccountID:         "123",
					Region:            "us-west-2",
					SslCertificateArn: "cert-arn",
				},
			}
			err := appConfig.Remote.ValidateFields()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	var _ = Describe("NewAppConfig", func() {
		It("should throw and error if app name is invalid", func() {
			appDir := helpers.GetTestApplicationDir("invalid-app-name")
			_, err := types.NewAppConfig(appDir)
			Expect(err).To(HaveOccurred())
			expectedErrorString := "The 'name' field 'invalid app' in application.yml is invalid. Only lowercase alphanumeric character(s) separated by a single hyphen are allowed. Must match regex: /^[a-z0-9]+(-[a-z0-9]+)*$/"
			Expect(err.Error()).To(ContainSubstring(expectedErrorString))
		})

		It("should throw and error if any service keys are invalid", func() {
			appDir := helpers.GetTestApplicationDir("invalid-app-service")
			_, err := types.NewAppConfig(appDir)
			Expect(err).To(HaveOccurred())
			expectedErrorString := "The service key 'services.invalid-service!' in application.yml is invalid. Only alphanumeric character(s) separated by a single hyphen are allowed. Must match regex: /^[a-zA-Z0-9]+(-[a-zA-Z0-9]+)*$/"
			Expect(err.Error()).To(ContainSubstring(expectedErrorString))
		})
	})

	var _ = Describe("GetAppConfig", func() {
		BeforeEach(func() {
			appDir := helpers.GetTestApplicationDir("complex-setup-app")
			var err error
			appConfig, err = types.NewAppConfig(appDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should include name, version and description", func() {
			Expect(appConfig.Name).To(Equal("complex-setup-app"))
			Expect(appConfig.Description).To(Equal("An app with complex setup used for testing"))
			Expect(appConfig.Version).To(Equal("0.0.1"))
		})

		It("should have all the dependencies", func() {
			Expect(appConfig.Local.Dependencies).To(Equal([]types.LocalDependency{
				types.LocalDependency{
					Name:    "exocom",
					Version: "0.26.1",
				},
				types.LocalDependency{
					Name:    "mongo",
					Version: "3.4.0",
					Config: types.LocalDependencyConfig{
						Ports:                 []string{"4000:4000"},
						Persist:               []string{"/data/db"},
						DependencyEnvironment: map[string]string{"DB_NAME": "test-db"},
						ServiceEnvironment:    map[string]string{"COLLECTION_NAME": "test-collection"},
					},
				},
			}))
		})

		It("should have all the services", func() {
			expectedServices := map[string]types.ServiceSource{
				"todo-service": types.ServiceSource{Location: "./todo-service"},
				"users-service": types.ServiceSource{
					MessageTranslations: []types.MessageTranslation{
						types.MessageTranslation{
							Public:   "users create",
							Internal: "mongo create",
						},
					},
					Location: "./users-service"},
				"html-server":      types.ServiceSource{Location: "./html-server"},
				"api-service":      types.ServiceSource{Location: "./api-service"},
				"external-service": types.ServiceSource{DockerImage: "originate/test-web-server:0.0.1"},
			}
			Expect(appConfig.Services).To(Equal(expectedServices))
		})
	})

	var _ = Describe("GetRemoteDependencyNames", func() {
		It("should return the names of all application dependencies", func() {
			appConfig = types.AppConfig{
				Remote: types.AppRemoteConfig{Dependencies: []types.RemoteDependency{
					{Name: "exocom"},
					{Name: "mongo"},
				},
				},
			}
			actual := appConfig.GetRemoteDependencyNames()
			expected := []string{"exocom", "mongo"}
			Expect(actual).To(Equal(expected))
		})
	})

	var _ = Describe("GetSortedServiceRoles", func() {
		It("should return the names of all services in alphabetical order", func() {
			appConfig = types.AppConfig{
				Services: map[string]types.ServiceSource{
					"worker-service-1": types.ServiceSource{},
					"public-service-1": types.ServiceSource{},
					"public-service-2": types.ServiceSource{},
				},
			}
			actual := appConfig.GetSortedServiceRoles()
			expected := []string{"public-service-1", "public-service-2", "worker-service-1"}
			Expect(actual).To(Equal(expected))
		})
	})

	var _ = Describe("VerifyServiceRoleDoesNotExist", func() {
		BeforeEach(func() {
			appConfig = types.AppConfig{
				Services: map[string]types.ServiceSource{
					"public-service-1": types.ServiceSource{},
					"worker-service-1": types.ServiceSource{},
				},
			}
		})

		It("should return error when the given service already exists", func() {
			err := appConfig.VerifyServiceRoleDoesNotExist("public-service-1")
			Expect(err).To(HaveOccurred())
			err = appConfig.VerifyServiceRoleDoesNotExist("worker-service-1")
			Expect(err).To(HaveOccurred())
		})

		It("should not return an error when the given service does not exist", func() {
			err := appConfig.VerifyServiceRoleDoesNotExist("new-service")
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
