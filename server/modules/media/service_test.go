package media

import (
	"io"
	"mime/multipart"
	"testing"
	"time"

	"zmessage/server/dal"
	"zmessage/server/models"
)

func setupTestService(t *testing.T) Service {
	t.Helper()

	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	return NewService(mgr, t.TempDir())
}

func createTestFile(content string, filename string) multipart.File {
	data := []byte(content)
	return &testFile{
		data:     data,
		filename:  filename,
		closeFlag: false,
	}
}

type testFile struct {
	data      []byte
	filename  string
	closeFlag bool
	offset    int64
}

func (f *testFile) Read(p []byte) (n int, err error) {
	if f.offset >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n = copy(p, f.data[f.offset:])
	f.offset += int64(n)
	return n, nil
}

func (f *testFile) ReadAt(p []byte, off int64) (int, error) {
	if off >= int64(len(f.data)) {
		return 0, io.EOF
	}
	n := copy(p, f.data[off:])
	return n, nil
}

func (f *testFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case 0:
		f.offset = offset
	case 1:
		f.offset += offset
	case 2:
		f.offset = int64(len(f.data)) + offset
	}
	if f.offset < 0 {
		f.offset = 0
	}
	if f.offset > int64(len(f.data)) {
		f.offset = int64(len(f.data))
	}
	return f.offset, nil
}

func (f *testFile) Close() error {
	f.closeFlag = true
	return nil
}

func TestService_ValidateType(t *testing.T) {
	svc := setupTestService(t)

	tests := []struct {
		filename string
		mimeType string
		wantType string
		wantErr  error
	}{
		{"test.jpg", "image/jpeg", "image", nil},
		{"test.png", "image/png", "image", nil},
		{"test.gif", "image/gif", "image", nil},
		{"test.mp3", "audio/mpeg", "voice", nil},
		{"test.wav", "audio/wav", "voice", nil},
		{"test.txt", "text/plain", "", ErrInvalidType},
		{"test.pdf", "application/pdf", "", ErrInvalidType},
		{"test", "", "", ErrInvalidType},
	}

	for _, tt := range tests {
		mediaType, err := svc.ValidateType(tt.filename, tt.mimeType)
		if tt.wantErr != nil {
			if err != tt.wantErr {
				t.Errorf("ValidateType(%q, %q) error = %v, want %v", tt.filename, tt.mimeType, err, tt.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("ValidateType(%q, %q) unexpected error: %v", tt.filename, tt.mimeType, err)
			continue
		}
		if mediaType != tt.wantType {
			t.Errorf("ValidateType(%q, %q) = %q, want %q", tt.filename, tt.mimeType, mediaType, tt.wantType)
		}
	}
}

func TestService_ValidateSize(t *testing.T) {
	svc := setupTestService(t)

	err := svc.ValidateSize(4*1024*1024, "image") // 4MB - 应该通过
	if err != nil {
		t.Errorf("4MB image should be valid, got: %v", err)
	}

	err = svc.ValidateSize(6*1024*1024, "image") // 6MB - 应该失败
	if err != ErrInvalidSize {
		t.Errorf("6MB image should be invalid, got: %v", err)
	}

	err = svc.ValidateSize(9*1024*1024, "voice") // 9MB - 应该通过
	if err != nil {
		t.Errorf("9MB voice should be valid, got: %v", err)
	}

	err = svc.ValidateSize(11*1024*1024, "voice") // 11MB - 应该失败
	if err != ErrInvalidSize {
		t.Errorf("11MB voice should be invalid, got: %v", err)
	}
}

func TestService_Delete(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	user := &models.User{
		Username:     "testuser",
		PasswordHash: "hash",
		Nickname:     "Test",
		CreatedAt:    time.Now().Unix(),
		LastSeen:     time.Now().Unix(),
	}
	if err := mgr.User().Create(user); err != nil {
		t.Fatalf("create user: %v", err)
	}

	media := &models.Media{
		OwnerID:     user.ID,
		Type:        "image",
		OriginalPath: "/tmp/test.jpg",
		Size:        1000,
		MimeType:     "image/jpeg",
		CreatedAt:    time.Now().Unix(),
	}
	if err := mgr.Media().Create(media); err != nil {
		t.Fatalf("create media: %v", err)
	}

	svc := NewService(mgr, t.TempDir())

	err = svc.Delete(media.ID, user.ID)
	if err != nil {
		t.Fatalf("delete own media failed: %v", err)
	}

	_, err = svc.Get(media.ID)
	if err != ErrNotFound {
		t.Error("media should be deleted")
	}
}

func TestService_GetWithInvalidID(t *testing.T) {
	svc := setupTestService(t)

	_, err := svc.Get(99999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}
}
