package types_test

import (
	"github.com/Originate/exosphere/src/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RemoteDependency", func() {

	Describe("validates required production fields", func() {
		It("throws an error if db-name is not valid", func() {
			missingConfig := types.RemoteDependency{
				Name:    "postgres",
				Version: "0.0.1",
				Config: types.RemoteDependencyConfig{
					Rds: types.RdsConfig{
						AllocatedStorage:   "10",
						DbName:             "test!",
						Username:           "test-user",
						PasswordSecretName: "TEST_PASSWORD",
						InstanceClass:      "db.t2.micro",
						StorageType:        "gp2",
						ServiceEnvVarNames: types.ServiceEnvVarNames{
							DbName:   "DB_NAME",
							Username: "DB_USER",
							Password: "DB_PASSWORD",
						},
					},
				},
			}
			err := missingConfig.ValidateFields()
			Expect(err).To(HaveOccurred())
			expectedErrorString := "production dependency postgres:0.0.1 has issues: only alphanumeric characters and hyphens allowed in 'rds.db-name'"
			Expect(err.Error()).To(ContainSubstring(expectedErrorString))
		})

		It("throws an error if production fields are missing", func() {
			missingConfig := types.RemoteDependency{
				Name:    "postgres",
				Version: "0.0.1",
				Config: types.RemoteDependencyConfig{
					Rds: types.RdsConfig{
						AllocatedStorage:   "10",
						Username:           "test-user",
						PasswordSecretName: "TEST_PASSWORD",
						InstanceClass:      "db.t2.micro",
						StorageType:        "gp2",
						ServiceEnvVarNames: types.ServiceEnvVarNames{
							DbName:   "DB_NAME",
							Username: "DB_USER",
							Password: "DB_PASSWORD",
						},
					},
				},
			}
			err := missingConfig.ValidateFields()
			Expect(err).To(HaveOccurred())
			expectedErrorString := "production dependency postgres:0.0.1 has issues: missing required field 'rds.db-name'"
			Expect(err.Error()).To(ContainSubstring(expectedErrorString))
		})

		It("does not throw an error if production fields are valid", func() {
			goodConfig := types.RemoteDependency{
				Name:    "postgres",
				Version: "0.0.1",
				Config: types.RemoteDependencyConfig{
					Rds: types.RdsConfig{
						AllocatedStorage:   "10",
						DbName:             "test",
						Username:           "test-user",
						PasswordSecretName: "TEST_PASSWORD",
						InstanceClass:      "db.t2.micro",
						StorageType:        "gp2",
						ServiceEnvVarNames: types.ServiceEnvVarNames{
							DbName:   "DB_NAME",
							Username: "DB_USER",
							Password: "DB_PASSWORD",
						},
					},
				},
			}
			err := goodConfig.ValidateFields()
			Expect(err).ToNot(HaveOccurred())
		})

	})
})
