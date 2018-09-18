package brats_test

import (
	"fmt"
	"github.com/BurntSushi/toml"
	libbuildpackV3 "github.com/buildpack/libbuildpack"
	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

var _ = Describe("Nodejs V3 buildpack", func() {
	It("should run V3 detection and build", func() {
		bpDir, err := cutlass.FindRoot()
		Expect(err).ToNot(HaveOccurred())

		workspaceDir, err := ioutil.TempDir("/tmp", "workspace")
		Expect(err).ToNot(HaveOccurred())
		fmt.Printf("WORKSPACE = %s", workspaceDir)
		//defer os.RemoveAll(workspaceDir)

		err = os.Chmod(workspaceDir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		appDir := filepath.Join(workspaceDir, "app")
		err = os.Mkdir(appDir, os.ModePerm)
		Expect(err).ToNot(HaveOccurred())

		err = libbuildpack.CopyDirectory(filepath.Join(bpDir, "fixtures", "simple_app"), appDir)
		Expect(err).ToNot(HaveOccurred())

		// We must ensure container cannot modify app dir
		err = os.Chmod(appDir, 0755)
		Expect(err).ToNot(HaveOccurred())

		// Run detect -----------------------------------------------------------------------------

		cmd := exec.Command(
			"docker",
			"run",
			"--rm",
			"-v",
			fmt.Sprintf("%s:/workspace", workspaceDir),
			"-v",
			fmt.Sprintf("%s:/buildpacks/%s/latest", bpDir, "org.cloudfoundry.buildpacks.nodejs"),
			os.Getenv("CNB_BUILD_IMAGE"),
			"/lifecycle/detector",
			"-order",
			"/buildpacks/org.cloudfoundry.buildpacks.nodejs/latest/fixtures/v3/order.toml",
			"-group",
			"/workspace/group.toml",
			"-plan",
			"/workspace/plan.toml",
		)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err = cmd.Run(); err != nil {
			Fail("failed to run V3 detection")
		}

		group := struct {
			Buildpacks []struct {
				Id      string `toml:"id"`
				Version string `toml:"version"`
			} `toml:"buildpacks"`
		}{}
		_, err = toml.DecodeFile(filepath.Join(workspaceDir, "group.toml"), &group)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(group.Buildpacks)).To(Equal(1))
		Expect(group.Buildpacks[0].Id).To(Equal("org.cloudfoundry.buildpacks.nodejs"))
		Expect(group.Buildpacks[0].Version).To(Equal("1.6.32"))

		plan := libbuildpackV3.BuildPlan{}
		_, err = toml.DecodeFile(filepath.Join(workspaceDir, "plan.toml"), &plan)
		Expect(err).ToNot(HaveOccurred())
		Expect(len(plan)).To(Equal(1))
		Expect(plan).To(HaveKey("node"))
		Expect(plan["node"].Version).To(Equal("~>10"))

		// Run build -----------------------------------------------------------------------------
		cmd = exec.Command(
			"docker",
			"run",
			"--rm",
			"-v",
			fmt.Sprintf("%s:/workspace", workspaceDir),
			"-v",
			fmt.Sprintf("%s:/buildpacks/%s/latest", bpDir, "org.cloudfoundry.buildpacks.nodejs"),
			os.Getenv("CNB_BUILD_IMAGE"),
			"/lifecycle/builder",
			"-group",
			"/workspace/group.toml",
			"-plan",
			"/workspace/plan.toml",
		)

		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			Fail("failed to run V3 build")
		}

		//launch := libbuildpackV3.LaunchMetadata{}
		//_, err = toml.DecodeFile(filepath.Join(workspaceDir, "launch", "launch.toml"), &launch)
		//Expect(err).ToNot(HaveOccurred())
	})
})
