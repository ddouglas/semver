package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const description = `semver [rc|p|m|M|h] [description]
	M for major
	m for minor
	p for patch
	rc to increment release candidate
	h for this message
	[description] tag description (optional)`

func main() {

	ctx := context.Background()

	for _, v := range []string{"git", "tail"} {
		if _, err := exec.LookPath(v); err != nil {
			panic(err)
		}
	}

	latestTag, err := getLatestTag(ctx)
	if err != nil {
		panic(err)
	}

	getOption(ctx, latestTag)

}

var validOpts = []string{"M", "m", "p", "rc", "q", "h"}

func getOption(ctx context.Context, currentTag string) {

	var opt string

	if len(os.Args) == 1 {
		fmt.Printf("Current Tag :: %s\n", currentTag)
		fmt.Println(description)
		var input string

	SCANLOOP:
		for {
			fmt.Scanln(&input)
			if validateOpt(input) {
				break SCANLOOP
			}

			fmt.Println("Invalid option, please try again")
		}

		opt = input
	}

	if len(os.Args) >= 2 {
		opt = os.Args[1]
		if !validateOpt(opt) {
			fmt.Println("Invalid option")
			os.Exit(1)
		}

	}

	parsedTag := parseTag(currentTag)

	switch opt {
	case "M":
		fmt.Println("Incrementing Major version")
		parsedTag.Major++
		parsedTag.Minor = 0
		parsedTag.Patch = 0
		parsedTag.RC = 0
	case "m":
		parsedTag.Minor++
		parsedTag.Patch = 0
		parsedTag.RC = 0
	case "p":
		if parsedTag.RC > 0 {
			parsedTag.RC = 0
		} else {
			parsedTag.Patch++
		}
	case "rc":
		parsedTag.RC++
	case "q":
		os.Exit(0)
	case "h":
		fmt.Println(description)
		os.Exit(0)
	default:
		fmt.Println("Invalid option")
	}

	tagDescription := ""

	if len(os.Args) == 3 {
		tagDescription = os.Args[2]
	}

	newTag := fmt.Sprintf("v%d.%d.%d", parsedTag.Major, parsedTag.Minor, parsedTag.Patch)
	if parsedTag.RC > 0 {
		newTag = fmt.Sprintf("%s-rc%d", newTag, parsedTag.RC)
	}

	fmt.Printf("%s --> %s. Confirm tag and push to origin? (description: %s) [y/n]: ", currentTag, newTag, tagDescription)
	// scanner that reads from stdin
	for {
		var input string
		fmt.Scanln(&input)
		if input == "y" {
			break
		} else if input == "n" {
			os.Exit(0)
		} else {
			fmt.Println("Invalid input. Please enter y or n")
		}
	}

	err := tag(ctx, newTag, tagDescription)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = push(ctx, newTag)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Tagged %s\n", newTag)

}

func validateOpt(opt string) bool {
	for _, v := range validOpts {
		if v == opt {
			return true
		}
	}
	return false
}

type Tag struct {
	Major int
	Minor int
	Patch int
	RC    int
}

func parseTag(tag string) Tag {

	tag = strings.TrimPrefix(tag, "v")

	parts := strings.Split(tag, ".")
	if len(parts) != 3 {
		fmt.Println("Invalid tag format")
		return Tag{
			Major: 0,
			Minor: 0,
			Patch: 0,
			RC:    0,
		}
	}

	if strings.Contains(parts[2], "-") {
		patchParts := strings.Split(parts[2], "-")
		parts[2] = patchParts[0]
		parts = append(parts, patchParts[1])
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		panic(err)
	}

	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		panic(err)
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		panic(err)
	}

	rc := 0

	if len(parts) == 4 {
		rcStr := strings.TrimPrefix(parts[3], "rc")

		rc, err = strconv.Atoi(rcStr)
		if err != nil {
			panic(err)
		}
	}

	return Tag{
		Major: major,
		Minor: minor,
		Patch: patch,
		RC:    rc,
	}

}

func getLatestTag(ctx context.Context) (string, error) {

	args := []string{
		"git",
		"tag",
		"-l",
		"'v[0-9]*.[0-9]*.[0-9]*'",
		"|",
		"tail",
		"-1",
	}

	argsStr := strings.Join(args, " ")

	cmd := exec.CommandContext(ctx, "/bin/bash", "-ec", argsStr)
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil

}

func tag(ctx context.Context, tag, descritpion string) error {

	args := []string{
		"git",
		"tag",
		"-a",
		tag,
	}

	if descritpion != "" {
		args = append(args, "-m", descritpion)
	}

	argsStr := strings.Join(args, " ")

	cmd := exec.CommandContext(ctx, "/bin/bash", "-ec", argsStr)
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error tagging: %s", string(out))
	}

	return nil

}

func push(ctx context.Context, tag string) error {

	args := []string{
		"git",
		"push",
		"origin",
		tag,
	}

	argsStr := strings.Join(args, " ")

	cmd := exec.CommandContext(ctx, "/bin/bash", "-ec", argsStr)
	cmd.Env = os.Environ()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error pushing tag: %s", string(out))
	}

	return nil

}
