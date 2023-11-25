package downloader

import (
	"errors"
	"path/filepath"
)

// PackType은 압축 파일의 타입을 나타냅니다.
type PackType string

const (
	// Zip은 zip 파일을 나타냅니다.
	Zip PackType = "zip"
	// Rar은 rar 파일을 나타냅니다.
	Rar PackType = "rar"
	// Tar은 tar 파일을 나타냅니다.
	Tar PackType = "tar"
	// TarGz은 tar.gz 파일을 나타냅니다.
	TarGz PackType = "tar.gz"
	// TarXz은 tar.xz 파일을 나타냅니다.
	TarXz PackType = "tar.xz"
	// TarBz2은 tar.bz2 파일을 나타냅니다.
	TarBz2 PackType = "tar.bz2"
	// 7z은 7z 파일을 나타냅니다.
	SevenZ PackType = "7z"
	// NotSupported은 지원하지 않는 파일을 나타냅니다.
	NotSupported PackType = "not_supported"
)

// UnpackerImpl은 downloader.Unpacker interface 의 구현체입니다.
type UnpackerImpl struct{}

// Unpack는 압축 파일을 풉니다.
func (u *UnpackerImpl) Unpack(filePath string, unpackPath string) error {
	// 압축 파일의 타입을 판별합니다.
	packType := getPackType(filePath)
	switch packType {
	case Zip:
		return unpackZip(filePath, unpackPath)
	case Rar:
		return unpackRar(filePath, unpackPath)
	case Tar:
		return unpackTar(filePath, unpackPath)
	case TarGz:
		return unpackTarGz(filePath, unpackPath)
	case TarXz:
		return unpackTarXz(filePath, unpackPath)
	case TarBz2:
		return unpackTarBz2(filePath, unpackPath)
	case SevenZ:
		return unpackSevenZ(filePath, unpackPath)
	case NotSupported:
		return errors.New("not supported pack type")
	}
	return nil
}

// getPackType은 압축 파일의 타입을 판별합니다.
func getPackType(filePath string) PackType {
	// 확장자를 추출합니다.
	ext := filepath.Ext(filePath)
	switch ext {
	case ".zip":
		return Zip
	case ".rar":
		return Rar
	case ".tar":
		return Tar
	case ".gz":
		return TarGz
	case ".xz":
		return TarXz
	case ".bz2":
		return TarBz2
	case ".7z":
		return SevenZ
	default:
		return NotSupported
	}
}

// unpackZip은 zip 파일을 풉니다.
func unpackZip(filePath string, unpackPath string) error {
	return nil
}

// unpackRar은 rar 파일을 풉니다.
func unpackRar(filePath string, unpackPath string) error {
	return nil
}

// unpackTar은 tar 파일을 풉니다.
func unpackTar(filePath string, unpackPath string) error {
	return nil
}

// unpackTarGz은 tar.gz 파일을 풉니다.
func unpackTarGz(filePath string, unpackPath string) error {
	return nil
}

// unpackTarXz은 tar.xz 파일을 풉니다.
func unpackTarXz(filePath string, unpackPath string) error {
	return nil
}

// unpackTarBz2은 tar.bz2 파일을 풉니다.
func unpackTarBz2(filePath string, unpackPath string) error {
	return nil
}

// unpackSevenZ은 7z 파일을 풉니다.
func unpackSevenZ(filePath string, unpackPath string) error {
	return nil
}
