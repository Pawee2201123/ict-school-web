package email

import (
	"fmt"
	"time"
)

// EnrollmentData contains information for the enrollment confirmation email
type EnrollmentData struct {
	StudentName  string
	ClassName    string
	RoomNumber   string
	RoomName     string
	TeacherName  string
	StartAt      time.Time
	EndAt        time.Time
}

// GenerateEnrollmentConfirmation creates the HTML body for enrollment confirmation
func GenerateEnrollmentConfirmation(data EnrollmentData) string {
	// Format times in Japanese style
	startDate := data.StartAt.Format("2006年01月02日")
	startTime := data.StartAt.Format("15:04")
	endTime := data.EndAt.Format("15:04")

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ja">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>授業申込完了</title>
</head>
<body style="font-family: 'メイリオ', Meiryo, 'ヒラギノ角ゴ Pro', 'Hiragino Kaku Gothic Pro', sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background-color: #f8f9fa; border-radius: 8px; padding: 30px; margin-bottom: 20px;">
        <h1 style="color: #0066cc; margin-top: 0;">授業申込完了のお知らせ</h1>
        <p><strong>%s</strong> 様</p>
        <p>模擬授業の申込が完了しました。以下の内容をご確認ください。</p>
    </div>

    <div style="background-color: #ffffff; border: 1px solid #dee2e6; border-radius: 8px; padding: 20px; margin-bottom: 20px;">
        <h2 style="color: #0066cc; border-bottom: 2px solid #0066cc; padding-bottom: 10px;">申込内容</h2>

        <table style="width: 100%%; border-collapse: collapse;">
            <tr>
                <td style="padding: 12px 0; border-bottom: 1px solid #dee2e6; font-weight: bold; width: 30%%;">授業名</td>
                <td style="padding: 12px 0; border-bottom: 1px solid #dee2e6;">%s</td>
            </tr>
            <tr>
                <td style="padding: 12px 0; border-bottom: 1px solid #dee2e6; font-weight: bold;">日時</td>
                <td style="padding: 12px 0; border-bottom: 1px solid #dee2e6;">%s %s 〜 %s</td>
            </tr>
            <tr>
                <td style="padding: 12px 0; border-bottom: 1px solid #dee2e6; font-weight: bold;">教室</td>
                <td style="padding: 12px 0; border-bottom: 1px solid #dee2e6;">%s %s</td>
            </tr>
            <tr>
                <td style="padding: 12px 0; font-weight: bold;">担当教員</td>
                <td style="padding: 12px 0;">%s</td>
            </tr>
        </table>
    </div>

    <div style="background-color: #fff3cd; border: 1px solid #ffc107; border-radius: 8px; padding: 15px; margin-bottom: 20px;">
        <p style="margin: 0; color: #856404;"><strong>⚠️ 注意事項</strong></p>
        <ul style="margin: 10px 0 0 0; padding-left: 20px; color: #856404;">
            <li>当日は開始時刻の10分前までにお越しください</li>
            <li>保護者の方もご一緒にご参加いただけます</li>
            <li>キャンセルされる場合は、お早めにご連絡ください</li>
        </ul>
    </div>

    <div style="background-color: #f8f9fa; border-radius: 8px; padding: 15px; text-align: center; font-size: 0.9em; color: #6c757d;">
        <p style="margin: 0;">このメールは送信専用です。ご返信いただいても対応できませんのでご了承ください。</p>
        <p style="margin: 5px 0 0 0;">お問い合わせは学校までご連絡ください。</p>
    </div>
</body>
</html>
`,
		data.StudentName,
		data.ClassName,
		startDate,
		startTime,
		endTime,
		data.RoomNumber,
		data.RoomName,
		data.TeacherName,
	)

	return html
}

// GetEnrollmentSubject returns the subject line for enrollment confirmation
func GetEnrollmentSubject() string {
	return "【模擬授業】申込完了のお知らせ"
}
