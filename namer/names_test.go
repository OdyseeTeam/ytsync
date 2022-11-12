package namer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_getClaimNameFromTitle(t *testing.T) {
	name := getClaimNameFromTitle("СтопХам - \"В ожидании ответа\"", 0)
	assert.Equal(t, "стопхам-в-ожидании", name)
	name = getClaimNameFromTitle("SADB - \"A Weak Woman With a Strong Hood\"", 0)
	assert.Equal(t, "sadb-a-weak-woman-with-a-strong-hood", name)
	name = getClaimNameFromTitle("錢包整理術 5 Tips、哪種錢包最NG？｜有錢人默默在做的「錢包整理術」 ft.@SHIN LI", 0)
	assert.Equal(t, "錢包整理術-5-tips-哪種錢包最ng", name)
	name = getClaimNameFromTitle("اسرع-طريقة-لتختيم", 0)
	assert.Equal(t, "اسرع-طريقة-لتختيم", name)
	name = getClaimNameFromTitle("شكرا على 380 مشترك😍😍😍😍 لي يريد دعم ادا وصلنا المقطع 40 لايك وراح ادعم قناتين", 0)
	assert.Equal(t, "شكرا-على-380-مشترك😍😍😍", name)
	name = getClaimNameFromTitle("test-@", 0)
	assert.Equal(t, "test", name)
	name = getClaimNameFromTitle("『あなたはただの空の殻でした』", 0)
	assert.Equal(t, "『あなたはただの空の殻でした』", name)
	name = getClaimNameFromTitle("精靈樂章-這樣的夥伴沒問題嗎 幽暗隕石坑（夢魘） 王有無敵狀態...要會閃不然會被秒（無課）", 2)
	assert.Equal(t, "精靈樂章-這樣的夥伴沒問題嗎-2", name)
	name = getClaimNameFromTitle("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", 50)
	assert.Equal(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-50", name)
	name = getClaimNameFromTitle("wtf*aaa", 0)
	assert.Equal(t, "wtf-aaa", name)
	name = getClaimNameFromTitle("wtf-*aaa", 0)
	assert.Equal(t, "wtf-aaa", name)
}
