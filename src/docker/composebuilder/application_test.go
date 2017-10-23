package composebuilder_test

import (
	"path"
	"runtime"
	"strings"

	"github.com/Originate/exosphere/src/docker/composebuilder"
	"github.com/Originate/exosphere/src/types"
	"github.com/Originate/exosphere/src/util"
	"github.com/Originate/exosphere/test_helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("composebuilder", func() {
	var _ = Describe("GetApplicationDockerConfigs", func() {
		var filePath string

		BeforeEach(func() {
			_, filePath, _, _ = runtime.Caller(0)
		})

		It("should return the proper docker configs for production", func() {
			err := testHelpers.CheckoutApp(cwd, "rds")
			Expect(err).NotTo(HaveOccurred())
			appDir := path.Join(path.Dir(filePath), "tmp", "rds")
			appConfig, err := types.NewAppConfig(appDir)
			Expect(err).NotTo(HaveOccurred())
			dockerConfigs, err := composebuilder.GetApplicationDockerConfigs(composebuilder.ApplicationOptions{
				AppConfig: appConfig,
				AppDir:    appDir,
				BuildMode: composebuilder.BuildModeDeployProduction,
				HomeDir:   homeDir,
			})
			Expect(err).NotTo(HaveOccurred())

			By("ignoring rds dependencies")
			delete(dockerConfigs, "my-sql-service")
			Expect(len(dockerConfigs)).To(Equal(0))
		})

		It("should return the proper docker configs for development", func() {
			err := testHelpers.CheckoutApp(cwd, "complex-setup-app")
			Expect(err).NotTo(HaveOccurred())
			internalServices := []string{"html-server", "todo-service", "users-service"}
			externalServices := []string{"external-service"}
			internalDependencies := []string{"exocom0.26.1"}
			externalDependencies := []string{"mongo3.4.0"}
			allServices := util.JoinStringSlices(internalServices, externalServices, internalDependencies, externalDependencies)
			appDir := path.Join(path.Dir(filePath), "tmp", "complex-setup-app")
			appConfig, err := types.NewAppConfig(appDir)
			Expect(err).NotTo(HaveOccurred())

			dockerConfigs, err := composebuilder.GetApplicationDockerConfigs(composebuilder.ApplicationOptions{
				AppConfig: appConfig,
				AppDir:    appDir,
				BuildMode: composebuilder.BuildModeLocalDevelopment,
				HomeDir:   homeDir,
			})
			Expect(err).NotTo(HaveOccurred())

			By("generate an image name for each dependency and external service")
			for _, serviceRole := range util.JoinStringSlices(internalDependencies, externalDependencies, externalServices) {
				Expect(len(dockerConfigs[serviceRole].Image)).ToNot(Equal(0))
			}

			By("should generate a container name for each service and dependency")
			for _, serviceRole := range allServices {
				Expect(len(dockerConfigs[serviceRole].ContainerName)).ToNot(Equal(0))
			}

			By("should have the correct build command for each internal service and dependency")
			for _, serviceRole := range internalServices {
				Expect(dockerConfigs[serviceRole].Command).To(Equal(`echo "does not run"`))
			}
			Expect(dockerConfigs["exocom0.26.1"].Command).To(Equal(""))

			By("should include 'exocom' in the dependencies of every service")
			for _, serviceRole := range append(internalServices, externalServices...) {
				exists := util.DoesStringArrayContain(dockerConfigs[serviceRole].DependsOn, "exocom0.26.1")
				Expect(exists).To(Equal(true))
			}

			By("should include external dependencies as dependencies")
			exists := util.DoesStringArrayContain(dockerConfigs["todo-service"].DependsOn, "mongo3.4.0")
			Expect(exists).To(Equal(true))

			By("should include the correct exocom environment variables")
			environment := dockerConfigs["exocom0.26.1"].Environment
			expectedServiceRoutes := []string{
				`{"receives":["todo.create"],"role":"todo-service","sends":["todo.created"]}`,
				`{"namespace":"mongo","receives":["mongo.list","mongo.create"],"role":"users-service","sends":["mongo.listed","mongo.created"]}`,
				`{"receives":["todo.created"],"role":"html-server","sends":["todo.create"]}`,
				`{"receives":["users.listed","users.created"],"role":"external-service","sends":["users.list","users.create"]}`,
			}
			for _, serviceRoute := range expectedServiceRoutes {
				Expect(strings.Contains(environment["SERVICE_ROUTES"], serviceRoute))
			}

			By("should include exocom environment variables in internal services' environment")
			for _, serviceRole := range internalServices {
				environment := dockerConfigs[serviceRole].Environment
				Expect(environment["EXOCOM_HOST"]).To(Equal("exocom0.26.1"))
			}

			By("should generate a volume path for an external dependency that mounts a volume")
			Expect(len(dockerConfigs["mongo3.4.0"].Volumes)).NotTo(Equal(0))

			By("should have the specified image and container names for the external service")
			serviceRole := "external-service"
			imageName := "originate/test-web-server"
			Expect(dockerConfigs[serviceRole].Image).To(Equal(imageName))
			Expect(dockerConfigs[serviceRole].ContainerName).To(Equal(serviceRole))

			By("should have the specified ports, volumes and environment variables for the external service")
			serviceRole = "external-service"
			ports := []string{"5000:5000"}
			Expect(dockerConfigs[serviceRole].Ports).To(Equal(ports))
			Expect(len(dockerConfigs[serviceRole].Volumes)).NotTo(Equal(0))
			Expect(dockerConfigs[serviceRole].Environment["EXTERNAL_SERVICE_HOST"]).To(Equal("external-service0.1.2"))
			Expect(dockerConfigs[serviceRole].Environment["EXTERNAL_SERVICE_PORT"]).To(Equal("$EXTERNAL_SERVICE_PORT"))

			By("should have the ports for the external dependency defined in application.yml")
			serviceRole = "mongo3.4.0"
			ports = []string{"4000:4000"}
			Expect(dockerConfigs[serviceRole].Ports).To(Equal(ports))
			Expect(len(dockerConfigs[serviceRole].Volumes)).NotTo(Equal(0))
		})
	})
})