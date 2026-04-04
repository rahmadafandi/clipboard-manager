//go:build windows

package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"golang.design/x/clipboard"
)

const psPopupScript = `
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$items = @($input)
if ($items.Count -eq 0) { exit 1 }
$form = New-Object System.Windows.Forms.Form
$form.Text = 'Clipboard History'
$form.Size = New-Object System.Drawing.Size(450, 350)
$form.StartPosition = 'CenterScreen'
$form.TopMost = $true
$listBox = New-Object System.Windows.Forms.ListBox
$listBox.Dock = 'Fill'
$listBox.Font = New-Object System.Drawing.Font('Consolas', 10)
$items | ForEach-Object { [void]$listBox.Items.Add($_) }
$listBox.SelectedIndex = 0
$listBox.Add_MouseDoubleClick({ $form.DialogResult = 'OK'; $form.Close() })
$listBox.Add_KeyDown({
    if ($_.KeyCode -eq 'Enter') { $form.DialogResult = 'OK'; $form.Close() }
    if ($_.KeyCode -eq 'Escape') { $form.Close() }
})
$form.Controls.Add($listBox)
$result = $form.ShowDialog()
if ($result -eq 'OK' -and $listBox.SelectedItem) {
    Write-Output $listBox.SelectedItem
}
`

var windowsLaunchers = []launcher{
	{
		bin: "powershell.exe",
		args: func() []string {
			return []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", psPopupScript}
		},
		imgArgs:  func() []string { return nil },
		fmtLine:  defaultFmtLine,
		parseIdx: func(output string) (int, bool) { return parseIdxFromNumber(output) },
	},
}

func detectLauncher() (*launcher, error) {
	for i := range windowsLaunchers {
		if _, err := exec.LookPath(windowsLaunchers[i].bin); err == nil {
			return &windowsLaunchers[i], nil
		}
	}
	return nil, fmt.Errorf("PowerShell not found")
}

func writeTextToClipboard(text string) {
	// Windows clipboard persists by default
	if err := clipboard.Init(); err == nil {
		clipboard.Write(clipboard.FmtText, []byte(text))
	}
}

func writeImageToClipboard(data []byte) {
	// Windows clipboard persists by default
	if err := clipboard.Init(); err == nil {
		clipboard.Write(clipboard.FmtImage, data)
	}
}

func showNotification(msg string) {
	script := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
$n = New-Object System.Windows.Forms.NotifyIcon
$n.Icon = [System.Drawing.SystemIcons]::Information
$n.Visible = $true
$n.ShowBalloonTip(2000, 'Clipboard Manager', '%s', 'Info')
Start-Sleep -Seconds 3
$n.Dispose()
`, strings.ReplaceAll(msg, "'", "''"))
	exec.Command("powershell.exe", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script).Start()
}
