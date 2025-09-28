package interactive

import (
	"github.com/joyfuldevs/project-lumos/pkg/slack"
	"github.com/joyfuldevs/project-lumos/pkg/slack/blockkit"
)

type ResponseType string

const (
	// 채널에 포함된 모두가 볼 수 있도록 메시지를 전달한다.
	InChannel ResponseType = "in_channel"
	// 상호작용한 사용자만 볼 수 있도록 메시지를 전달한다.
	Ephemeral ResponseType = "ephemeral"
)

// 이벤트에 대한 응답으로 전달하는 데이터.
type ResponsePayload struct {
	// 응답이 채널에 포함된 모두가 볼 수 있도록 할지, 상호작용한 사용자만 볼 수 있도록 할지 지정합니다.
	// 기본값은 Ephemeral 입니다.
	ResponseType ResponseType `json:"response_type,omitempty"`
	// 블록을 사용하는 경우에는 알림에 표시할 대체 문자열로 사용됩니다.
	// 그렇지 않은 경우에는 메시지의 본문 텍스트가 됩니다.
	//
	// 이 필드는 블록 사용 시 필수 입력 사항은 아니지만 대체 문자열로 포함할 것을 권장합니다.
	Text string `json:"text"`
	// Block Kit 블록 배열입니다.
	Blocks []*blockkit.Block `json:"blocks,omitempty"`
	// 메시지를 스레드에 게시하려면 이 필드에 스레드 타임스탬프를 포함합니다.
	//
	// Bot과의 대화는 스레드 형태이므로 이 필드를 포함해야 합니다.
	// 이 필드를 포함하지 않으면 메시지가 새 채팅 메시지로 게시됩니다.
	ThreadTimestamp slack.Timestamp `json:"thread_ts,omitempty"`
	// 원본 메시지의 삭제 여부를 지정합니다.
	DeleteOriginal bool `json:"delete_original"`
	// 원본 메시지의 교체 여부를 지정합니다.
	ReplaceOriginal bool `json:"replace_original"`
}
