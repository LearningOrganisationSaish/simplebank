package mail

import (
	"github.com/SaishNaik/simplebank/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip("skip, please run without -short")
	}
	config, err := utils.LoadConfig("..") // parent folder of the mail folder
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	subject := "A test email"
	content := `
		<h1>A test email</h1>
		<p>This is a test message from simple bank</p>
	`

	to := []string{
		config.EmailSenderAddress,
	}
	attachFiles := []string{"../README.md"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
