package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Note struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Text      string    `json:"text"`
}

type MessageResponse struct {
	Messages []Note `json:"messages"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println(".envファイルが読み取れないか存在しません。")
		return
	}

	// 消去したい単語をenvで読み取ってもらうつもりなんだけど、どんな感じで読み取らせようか。
	misskeyHost := os.Getenv("MISSKEY_HOST")
	misskeyToken := os.Getenv("MISSKEY_TOKEN")
	targetString := os.Getenv("TARGET_STRING")

	// 取得した投稿をフィルタリング
	filteredNotes, err := fetchAndFilterNotes(misskeyHost, misskeyToken, targetString)
	if err != nil {
		fmt.Println("フィルタリングに失敗しました。:", err)
		return
	}

	for _, note := range filteredNotes {
		err := deleteNote(misskeyHost, misskeyToken, note.ID)
		if err != nil {
			fmt.Printf("何故か消えてくれませんでした。 %s: %v\n", note.ID, err)
		} else {
			fmt.Printf("ぼん！%s 粛清しました。\n", note.ID)
		}
	}
}

// グローバルタイムラインから投稿を取得して指定した文字列を含む投稿をフィルタリングする
// ここをどうするか未定
func fetchAndFilterNotes(apiURL, authToken, targetString string) ([]Note, error) {
	notes, err := getGlobalTimelineNotes(apiURL, authToken)
	if err != nil {
		return nil, err
	}

	filteredNotes := filterNotes(notes, targetString)

	return filteredNotes, nil
}

// 試しに特定の文字列を含む投稿をフィルタリングする
func filterNotes(notes []Note, targetString string) []Note {
	filtered := make([]Note, 0)

	for _, note := range notes {
		if containsTargetString(note.Text, targetString) {
			filtered = append(filtered, note)
		}
	}

	return filtered
}

// ここDocみてそのまんま書いたから動かない
func getGlobalTimelineNotes(apiURL, authToken string) ([]Note, error) {
	url := fmt.Sprintf("https://%s/api/notes/global-timeline", apiURL)
	// ここPOSTって書いてあったんだけどマジで？
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var messageResponse MessageResponse
	err = json.Unmarshal(body, &messageResponse)
	if err != nil {
		return nil, err
	}

	return messageResponse.Messages, nil
}

func deleteNote(apiURL, authToken, noteID string) error {
	url := fmt.Sprintf("https://%s/api/notes/delete", apiURL)

	reqBody, err := json.Marshal(map[string]string{
		"noteId": noteID,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", authToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("エラー。: %d", resp.StatusCode)
	}

	return nil
}

func containsTargetString(text, targetString string) bool {
	return strings.Contains(text, targetString)
}
