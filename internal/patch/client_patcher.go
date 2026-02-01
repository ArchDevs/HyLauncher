package patch

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"HyLauncher/internal/env"
	"HyLauncher/internal/progress"
	"HyLauncher/pkg/fileutil"
	"HyLauncher/pkg/model"
)

const (
	originalDomain   = "hytale.com"
	minDomainLength  = 4
	maxDomainLength  = 10
	defaultNewDomain = "sanasol.ws"
)

type ClientPatcher struct {
	targetDomain string
}

func NewClientPatcher(targetDomain string) *ClientPatcher {
	if len(targetDomain) < minDomainLength || len(targetDomain) > maxDomainLength {
		fmt.Printf("Warning: Domain %q is invalid (must be %d-%d chars), using default: %s\n",
			targetDomain, minDomainLength, maxDomainLength, defaultNewDomain)
		targetDomain = defaultNewDomain
	}
	return &ClientPatcher{
		targetDomain: targetDomain,
	}
}

// StringToLengthPrefixed converts a string to length-prefixed byte format
// Format: [length byte] [00 00 00 padding] [char1] [00] [char2] [00] ... [lastChar]
// Note: No null byte after the last character
func (cp *ClientPatcher) StringToLengthPrefixed(str string) []byte {
	length := len(str)
	result := make([]byte, 4+length+(length-1))

	result[0] = byte(length)
	// bytes 1-3 stay zero (padding)

	pos := 4
	for i := 0; i < length; i++ {
		result[pos] = str[i]
		pos++
		if i < length-1 {
			result[pos] = 0x00
			pos++
		}
	}

	return result
}

// StringToUTF16LE converts a string to UTF-16LE bytes (for dotnet)
func (cp *ClientPatcher) StringToUTF16LE(str string) []byte {
	buf := make([]byte, len(str)*2)
	for i, r := range str {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(r))
	}
	return buf
}

// StringToUTF8 converts a string to UTF-8 bytes (for Java)
func (cp *ClientPatcher) StringToUTF8(str string) []byte {
	return []byte(str)
}

// FindAllOccurrences finds all positions of pattern in data
func (cp *ClientPatcher) FindAllOccurrences(data, pattern []byte) []int {
	var positions []int
	pos := 0

	for pos < len(data) {
		index := bytes.Index(data[pos:], pattern)
		if index == -1 {
			break
		}
		positions = append(positions, pos+index)
		pos = pos + index + 1
	}

	return positions
}

// ReplaceBytes replaces all occurrences of oldBytes with newBytes
func (cp *ClientPatcher) ReplaceBytes(data, oldBytes, newBytes []byte) ([]byte, int) {
	if len(newBytes) > len(oldBytes) {
		fmt.Printf("  Warning: New pattern (%d) longer than old (%d), skipping\n", len(newBytes), len(oldBytes))
		return data, 0
	}

	result := make([]byte, len(data))
	copy(result, data)

	positions := cp.FindAllOccurrences(result, oldBytes)

	for _, pos := range positions {
		copy(result[pos:], newBytes)
		for i := len(newBytes); i < len(oldBytes); i++ {
			result[pos+i] = 0x00
		}
	}

	return result, len(positions)
}

// FindAndReplaceDomainUTF8 replaces domain in UTF-8 format (for Java JARs)
func (cp *ClientPatcher) FindAndReplaceDomainUTF8(data []byte, oldDomain, newDomain string) ([]byte, int) {
	result := make([]byte, len(data))
	copy(result, data)

	oldUtf8 := cp.StringToUTF8(oldDomain)
	newUtf8 := cp.StringToUTF8(newDomain)

	positions := cp.FindAllOccurrences(result, oldUtf8)
	count := 0

	for _, pos := range positions {
		if pos+len(oldUtf8) <= len(result) && len(newUtf8) <= len(oldUtf8) {
			copy(result[pos:], newUtf8)
			for i := len(newUtf8); i < len(oldUtf8); i++ {
				result[pos+i] = 0x00
			}
			count++
			fmt.Printf("  Patched UTF-8 occurrence at offset 0x%x\n", pos)
		}
	}

	return result, count
}

// FindAndReplaceDomainSmart handles both null-terminated and length-prefixed UTF-16LE strings
func (cp *ClientPatcher) FindAndReplaceDomainSmart(data []byte, oldDomain, newDomain string) ([]byte, int) {
	result := make([]byte, len(data))
	copy(result, data)

	if len(newDomain) > len(oldDomain) {
		fmt.Printf("  Skipping UTF-16LE patch: new domain too long (%d > %d)\n", len(newDomain), len(oldDomain))
		return result, 0
	}

	oldUtf16NoLast := cp.StringToUTF16LE(oldDomain[:len(oldDomain)-1])
	newUtf16NoLast := cp.StringToUTF16LE(newDomain[:len(newDomain)-1])
	oldLastCharByte := byte(oldDomain[len(oldDomain)-1])
	newLastCharByte := byte(newDomain[len(newDomain)-1])

	positions := cp.FindAllOccurrences(result, oldUtf16NoLast)
	count := 0

	for _, pos := range positions {
		lastCharPos := pos + len(oldUtf16NoLast)
		if lastCharPos+1 > len(result) {
			continue
		}

		if result[lastCharPos] != oldLastCharByte {
			continue
		}

		copy(result[pos:], newUtf16NoLast)
		result[lastCharPos] = newLastCharByte

		if len(newDomain) < len(oldDomain) {
			paddingStart := pos + len(cp.StringToUTF16LE(newDomain))
			paddingEnd := pos + len(cp.StringToUTF16LE(oldDomain))
			for i := paddingStart; i < paddingEnd && i < len(result); i++ {
				result[i] = 0x00
			}
		}

		if lastCharPos+1 < len(result) && result[lastCharPos+1] == 0x00 {
			fmt.Printf("  Patched UTF-16LE occurrence at offset 0x%x\n", pos)
		} else {
			fmt.Printf("  Patched length-prefixed occurrence at offset 0x%x\n", pos)
		}
		count++
	}

	return result, count
}

// ApplyDomainPatches applies all domain patches to binary data
func (cp *ClientPatcher) ApplyDomainPatches(data []byte, protocol string) ([]byte, int) {
	result := make([]byte, len(data))
	copy(result, data)
	totalCount := 0

	fmt.Printf("  Replacing %s -> %s\n", originalDomain, cp.targetDomain)

	var count int

	// Patch sentry URL
	oldSentry := "https://ca900df42fcf57d4dd8401a86ddd7da2@sentry.hytale.com/2"
	newSentry := fmt.Sprintf("%st@%s/2", protocol, cp.targetDomain)
	result, count = cp.ReplaceBytes(result, cp.StringToLengthPrefixed(oldSentry), cp.StringToLengthPrefixed(newSentry))
	if count > 0 {
		fmt.Printf("  Patched sentry URL: %d occurrences\n", count)
		totalCount += count
	}

	// Patch full URLs with subdomains
	subdomains := []string{"", "sessions.", "account-data.", "gameservers."}
	for _, sub := range subdomains {
		oldUrl := fmt.Sprintf("%s%s%s", protocol, sub, originalDomain)
		newUrl := fmt.Sprintf("%s%s", protocol, cp.targetDomain)
		result, count = cp.ReplaceBytes(result, cp.StringToLengthPrefixed(oldUrl), cp.StringToLengthPrefixed(newUrl))
		if count > 0 {
			fmt.Printf("  Patched URL %s: %d occurrences\n", oldUrl, count)
			totalCount += count
		}
	}

	// Patch bare domain
	result, count = cp.ReplaceBytes(result, cp.StringToLengthPrefixed(originalDomain), cp.StringToLengthPrefixed(cp.targetDomain))
	if count > 0 {
		fmt.Printf("  Patched bare domain: %d occurrences\n", count)
		totalCount += count
	}

	// Patch UTF-16LE strings
	result, count = cp.FindAndReplaceDomainSmart(result, originalDomain, cp.targetDomain)
	if count > 0 {
		fmt.Printf("  Patched UTF-16LE: %d occurrences\n", count)
		totalCount += count
	}

	return result, totalCount
}

// PatchClient patches the client binary
func (cp *ClientPatcher) PatchClient(clientPath string, reporter *progress.Reporter) error {
	fmt.Printf("=== Client Patcher ===\n")
	fmt.Printf("Target: %s\n", clientPath)
	fmt.Printf("Domain: %s (%d chars)\n", cp.targetDomain, len(cp.targetDomain))

	if !fileutil.FileExists(clientPath) {
		return fmt.Errorf("client binary not found: %s", clientPath)
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 10, "Reading client binary...")
	}

	data, err := os.ReadFile(clientPath)
	if err != nil {
		return fmt.Errorf("failed to read client: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 30, "Applying patches...")
	}

	patchedData, count := cp.ApplyDomainPatches(data, "https://")

	if count == 0 {
		fmt.Println("No patches applied (already patched or no matches found)")
		return nil
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 70, "Writing patched binary...")
	}

	// Backup original
	backupPath := clientPath + ".original"
	if !fileutil.FileExists(backupPath) {
		if err := os.Rename(clientPath, backupPath); err != nil {
			return fmt.Errorf("failed to backup original: %w", err)
		}
	}

	// Write patched version
	if err := os.WriteFile(clientPath, patchedData, 0755); err != nil {
		return fmt.Errorf("failed to write patched client: %w", err)
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 100, fmt.Sprintf("Client patched (%d occurrences)", count))
	}

	fmt.Printf("Client patched successfully (%d total occurrences)\n", count)
	return nil
}

// PatchServer patches the server JAR file
func (cp *ClientPatcher) PatchServer(serverPath string, reporter *progress.Reporter) error {
	fmt.Printf("=== Server Patcher ===\n")
	fmt.Printf("Target: %s\n", serverPath)
	fmt.Printf("Domain: %s (%d chars)\n", cp.targetDomain, len(cp.targetDomain))

	if !fileutil.FileExists(serverPath) {
		return fmt.Errorf("server JAR not found: %s", serverPath)
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 10, "Opening server JAR...")
	}

	zipReader, err := zip.OpenReader(serverPath)
	if err != nil {
		return fmt.Errorf("failed to open JAR: %w", err)
	}
	defer zipReader.Close()

	tempPath := serverPath + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	zipWriter := zip.NewWriter(tempFile)
	defer zipWriter.Close()

	totalCount := 0
	oldUtf8 := cp.StringToUTF8(originalDomain)

	if reporter != nil {
		reporter.Report(progress.StagePatch, 30, "Patching JAR entries...")
	}

	for i, entry := range zipReader.File {
		if reporter != nil && i%100 == 0 {
			pct := 30 + (40 * i / len(zipReader.File))
			reporter.Report(progress.StagePatch, float64(pct), "Patching JAR entries...")
		}

		rc, err := entry.Open()
		if err != nil {
			return fmt.Errorf("failed to read entry %s: %w", entry.Name, err)
		}

		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to read entry data %s: %w", entry.Name, err)
		}

		shouldPatch := strings.HasSuffix(entry.Name, ".class") ||
			strings.HasSuffix(entry.Name, ".properties") ||
			strings.HasSuffix(entry.Name, ".json") ||
			strings.HasSuffix(entry.Name, ".xml") ||
			strings.HasSuffix(entry.Name, ".yml")

		if shouldPatch && bytes.Contains(data, oldUtf8) {
			patchedData, count := cp.FindAndReplaceDomainUTF8(data, originalDomain, cp.targetDomain)
			if count > 0 {
				data = patchedData
				totalCount += count
			}
		}

		writer, err := zipWriter.CreateHeader(&entry.FileHeader)
		if err != nil {
			return fmt.Errorf("failed to create entry %s: %w", entry.Name, err)
		}
		if _, err := writer.Write(data); err != nil {
			return fmt.Errorf("failed to write entry %s: %w", entry.Name, err)
		}
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 80, "Finalizing patched JAR...")
	}

	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close ZIP writer: %w", err)
	}
	tempFile.Close()
	zipReader.Close()

	if totalCount > 0 {
		backupPath := serverPath + ".original"
		if !fileutil.FileExists(backupPath) {
			if err := os.Rename(serverPath, backupPath); err != nil {
				os.Remove(tempPath)
				return fmt.Errorf("failed to backup original: %w", err)
			}
		} else {
			os.Remove(serverPath)
		}

		if err := os.Rename(tempPath, serverPath); err != nil {
			return fmt.Errorf("failed to replace with patched version: %w", err)
		}
	} else {
		os.Remove(tempPath)
		fmt.Println("No patches applied (already patched or no matches found)")
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 100, fmt.Sprintf("Server patched (%d occurrences)", totalCount))
	}

	fmt.Printf("Server patched successfully (%d total occurrences)\n", totalCount)
	return nil
}

// EnsureGamePatched ensures both client and server are patched
func EnsureGamePatched(ctx context.Context, request model.InstanceModel, targetDomain string, reporter *progress.Reporter) error {
	patcher := NewClientPatcher(targetDomain)

	// Patch client
	clientPath := env.GetGameClientPath(request.Branch, request.BuildVersion)
	if clientPath != "" {
		if reporter != nil {
			reporter.Report(progress.StagePatch, 0, "Patching client binary...")
		}

		if err := patcher.PatchClient(clientPath, reporter); err != nil {
			return fmt.Errorf("failed to patch client: %w", err)
		}
	} else {
		fmt.Println("Warning: Client binary not found, skipping client patch")
	}

	// Patch server
	serverPath := env.GetServerPath(request.Branch, request.BuildVersion)
	if serverPath != "" {
		if reporter != nil {
			reporter.Report(progress.StagePatch, 50, "Patching server JAR...")
		}

		if err := patcher.PatchServer(serverPath, reporter); err != nil {
			return fmt.Errorf("failed to patch server: %w", err)
		}
	} else {
		fmt.Println("Warning: Server JAR not found, skipping server patch")
	}

	if reporter != nil {
		reporter.Report(progress.StagePatch, 100, "Game patching complete")
	}

	return nil
}

func RestoreOriginalGame(request model.InstanceModel) error {
	restored := 0

	clientPath := env.GetGameClientPath(request.Branch, request.BuildVersion)
	if clientPath != "" {
		backupPath := clientPath + ".original"
		if fileutil.FileExists(backupPath) {
			if err := os.Remove(clientPath); err != nil {
				return fmt.Errorf("failed to remove patched client: %w", err)
			}
			if err := os.Rename(backupPath, clientPath); err != nil {
				return fmt.Errorf("failed to restore original client: %w", err)
			}
			fmt.Println("Restored original client binary")
			restored++
		}
	}

	serverPath := env.GetServerPath(request.Branch, request.BuildVersion)
	if serverPath != "" {
		backupPath := serverPath + ".original"
		if fileutil.FileExists(backupPath) {
			if err := os.Remove(serverPath); err != nil {
				return fmt.Errorf("failed to remove patched server: %w", err)
			}
			if err := os.Rename(backupPath, serverPath); err != nil {
				return fmt.Errorf("failed to restore original server: %w", err)
			}
			fmt.Println("Restored original server JAR")
			restored++
		}
	}

	if restored == 0 {
		return fmt.Errorf("no backups found to restore")
	}

	fmt.Printf("Restored %d original file\n", restored)
	return nil
}
