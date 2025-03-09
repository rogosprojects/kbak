package utils

// Terminal color codes
const (
	// Colors
	Reset   = "\033[0m"
	Black   = "\033[30m"
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	White   = "\033[37m"

	// Bright colors
	BrightBlack   = "\033[90m"
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
	BrightWhite   = "\033[97m"

	// Text formatting
	Bold      = "\033[1m"
	Underline = "\033[4m"
)

// Emojis for different message types
const (
	InfoEmoji    = "ℹ "
	SuccessEmoji = "✔ "
	WarningEmoji = "⚠️"
	ErrorEmoji   = "❌"
	BackupEmoji  = "💾"
	K8sEmoji     = "🚢"
	StartEmoji   = "🚀"
	DebugEmoji   = "🔍"
	SkippedEmoji = "⏭️"
)
