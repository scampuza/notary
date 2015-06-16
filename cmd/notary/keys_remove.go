package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	subjectKeyID  string
	cmdKeysRemove = &cobra.Command{
		Use:   "remove [ Subject Key ID ]",
		Short: "removes trust from a specific certificate authority or certificate.",
		Long:  "remove trust from a specific certificate authority.",
		Run:   keysRemove,
	}
)

func keysRemove(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.Usage()
		fatalf("must specify a SHA256 SubjectKeyID of the certificate")
	}

	cert, err := caStore.GetCertificateBySKID(args[0])
	if err != nil {
		fatalf("certificate not found")
	}

	fmt.Printf("Removing: ")
	printCert(cert)

	err = caStore.RemoveCert(cert)
	if err != nil {
		fatalf("failed to remove certificate for Key Store")
	}
}