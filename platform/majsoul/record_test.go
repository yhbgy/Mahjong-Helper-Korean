package majsoul

import (
	"os"
	"testing"
)

func TestDownloadRecords(t *testing.T) {
	username, ok := os.LookupEnv("USERNAME")
	if !ok {
		t.Skip("환경 변수 USERNAME이 설정되지 않아 종료합니다")
	}

	password, ok := os.LookupEnv("PASSWORD")
	if !ok {
		t.Skip("환경 변수 PASSWORD가 설정되지 않아 종료합니다")
	}

	if err := DownloadRecords(username, password, RecordTypeAll); err != nil {
		t.Fatal(err)
	}
}
