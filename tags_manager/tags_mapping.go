package tags_manager

import (
	"regexp"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	Gaming     = "gaming"
	Blockchain = "blockchain"
	News       = "news"
	Learning   = "learning"
	Funny      = "funny"
	Technology = "technology"
	Automotive = "automotive"
	Economics  = "economics"
	Sports     = "sports"
	Food       = "food"
	Science    = "science"
	Art        = "art"
	Nature     = "nature"
	Beliefs    = "beliefs"
	Music      = "music"
	PopCulture = "pop culture"
	Weapons    = "weapons"
)

func GetTagsForChannel(channelID string) []string {
	tags, _ := channelWideTags[channelID]
	return tags
}

func SanitizeTags(tags []string, youtubeChannelID string) ([]string, error) {
	unsanitized := make([]string, 0, len(tags))
	for _, t := range tags {
		t, err := normalizeTag(t)
		if err != nil {
			return nil, err
		}
		if t == "" {
			continue
		}
		unsanitized = append(unsanitized, t)
	}
	ts := &tagsSanitizer{
		Unsanitized: unsanitized,
		ChannelID:   youtubeChannelID,
	}
	ts.init()
	ts.cleanup()
	ts.replace()
	ts.add()

	originalTags := make([]string, 0, len(ts.Sanitized))
	curatedTags := make([]string, 0, len(ts.Sanitized))
	for t, curated := range ts.Sanitized {
		if curated {
			curatedTags = append(curatedTags, t)
		} else {
			originalTags = append(originalTags, t)
		}
	}
	sanitizedTags := make([]string, 0, len(originalTags)+len(curatedTags))
	sort.Strings(curatedTags)
	sort.Strings(originalTags)
	sanitizedTags = append(sanitizedTags, curatedTags...)
	sanitizedTags = append(sanitizedTags, originalTags...)
	return sanitizedTags, nil
}

const TagMaxLength = 50

func normalizeTag(t string) (string, error) {
	t = strings.ToLower(t)
	multipleSpaces := regexp.MustCompile(`\s{2,}`)
	leadingAndTrailingSpaces := regexp.MustCompile(`^\s+|\s$`)
	hashTags := regexp.MustCompile(`(#\d+\s)|#+`)
	inParenthesis := regexp.MustCompile(`\([^\)]+\)`)
	weirdChars := regexp.MustCompile(`[^-\w'& +\/A-Za-zÀ-ÖØ-öø-ÿ]`)
	startsOrEndsInWeirdChars := regexp.MustCompile(`^[^A-Za-zÀ-ÖØ-öø-ÿ0-9]+|[^A-Za-zÀ-ÖØ-öø-ÿ0-9]+$`)
	t = hashTags.ReplaceAllString(t, "")
	t = inParenthesis.ReplaceAllString(t, " ")
	t = startsOrEndsInWeirdChars.ReplaceAllString(t, "")
	t = multipleSpaces.ReplaceAllString(t, " ")
	t = leadingAndTrailingSpaces.ReplaceAllString(t, "")
	if weirdChars.MatchString(t) {
		log.Debugf("tag '%s' has weird stuff in it, skipping\n", t)
		return "", nil
	}
	if len(t) > TagMaxLength {
		log.Debugf("tag '%s' is too long, skipping\n", t)
		return "", nil
	}
	return t, nil
}

type tagsSanitizer struct {
	Unsanitized []string
	Sanitized   map[string]bool
	ChannelID   string
}

func (ts *tagsSanitizer) init() {
	if len(ts.Sanitized) == 0 {
		ts.Sanitized = make(map[string]bool, len(ts.Unsanitized)+16)
	}
}

func (ts *tagsSanitizer) cleanup() {
	for _, t := range ts.Unsanitized {
		_, ok := tagsToSkip[t]
		if !ok {
			ts.Sanitized[t] = false
		}
	}
}

func (ts *tagsSanitizer) replace() {
	for _, t := range ts.Unsanitized {
		match, filterMatch := mapAndReplace[t]
		if filterMatch {
			delete(ts.Sanitized, t)
			ts.Sanitized[match] = true
		}
	}
}

func (ts *tagsSanitizer) add() {
	for _, t := range ts.Unsanitized {
		match, filterMatch := mapAndKeep[t]
		if filterMatch {
			ts.Sanitized[match] = true
			ts.Sanitized[t] = false
		}
	}
	extraTags, ok := channelWideTags[ts.ChannelID]
	if ok {
		for _, t := range extraTags {
			ts.Sanitized[t] = true
		}
	}
}

const (
	SwissExperiments                    = "UCNQfQvFMPnInwsU_iGYArJQ"
	JustJuggling                        = "UCftqelpjmbFrUwr3VVzzVwA"
	anupjanakpur                        = "UC_5tRfC4L2AbTz6mj6vZrKw"
	PraveenMedia                        = "UC_fjE70lKNwM9AKIofQv-bA"
	MisteriosDesvelados                 = "UC-FzxivscjYzonBDXgX2GLg"
	kaipenchery                         = "UC-MU4K3Ghl-IdEX4J68tnTA"
	Dhinchakpooja                       = "UC-stzLwoQF_R8Vnfb28d7Lg"
	Karolajn                            = "UC-vYDJQiabg9BK8XTDPFspg"
	thefantasio974                      = "UC0GHeUcPxfNEZqbFyxy2frQ"
	EduPrimitivo                        = "UC0odmP6ffEw3iVwTEyt-UKA"
	isupportanupriyapatel               = "UC0sxOzdHnmauMdOra0cxXVg"
	theconfusedindian                   = "UC12ZPYxQbMA1loZE1OA609w"
	guriaarts                           = "UC1fHp166o1Hd024fg3-d-jg"
	minutodafisica                      = "UC1lUEcHrQwbusC5ext1himg"
	khabir                              = "UC1yzNCXXk1h_tHgsqChE8mw"
	OMapadaMinaMara                     = "UC2WMUPTbxQQ9hXq2_uutspg"
	oliverjanich                        = "UC3cmEfpy4XED7YYEe69nIMA"
	shyamsadhu                          = "UC3XAT8oBjL2RaqfblESHW_A"
	EbaduRahmantech                     = "UC4950gpY6Qw1lCAcN6gNWwQ"
	_1975oles                           = "UC4bGQWN4C8idymxlZxw77oA"
	_8dsongsbollywood                   = "UC4nH7zmw41lRDbaiIddumjw"
	EYESTV                              = "UC50CIbyHMydEEzEhV1ZNEBw"
	Nono                                = "UC5yufoRPJy-1e1pR73UTxFQ"
	lichtle                             = "UC6n_2v7YjwZ65F9h77nZPOA"
	elabuelokraken                      = "UC7iWk2xziMR4hs_NzfsYytw"
	cidvela                             = "UCaVOx5GCcSi2ELjWUdf05PA"
	jayaskumar                          = "UCax9CJCQ6aY0bntP3mcIWKQ"
	canaldarippi                        = "UCB_fba7yYMwa91F7rijBsVA"
	FunMusicClassical                   = "UCB_X256IN7QiBtDkueo_6QQ"
	minutodaterra                       = "UCB0zinWfy-dS_NqcOINYo3A"
	franciscoalves                      = "UCbAm6YcZGk04obnFb-LuOSA"
	criptomonedastv                     = "UCbK6Awel1-o-9JDVMFq808Q"
	lux77                               = "UCBYm4l3NX352goFtjSgZ7jA"
	MrLuckyOficial                      = "UCEbMhGhZ1JoOYKgbgVyeQaA"
	KotneKit                            = "UCf_1BQz3T151Eph0Sarj3Yw"
	jaysgaming                          = "UCFAkedtc3jjDZQqqIULxqWg"
	iamdineshthakur                     = "UCg0VInmUZoSdHD-DNmjifBA"
	Musicreationz                       = "UCGJEtZB0Nj3rT3zZfyfjIZQ"
	lafenixtv                           = "UCHAx8o0jj-8L_sqR5QM1Cwg"
	guardeiafe                          = "UCHWjmG-Ce93VNcNZMMUyGQQ"
	KHANEKIKHUSHBOO                     = "UCI8DRKcrfrHklRqSJHPhOwQ"
	AjayGamingYT                        = "UCi9r5igvQblbI0PhJFj-IDg"
	SzekelyVegan                        = "UCiqvKH1Ib_VPwh9bQf_hlXQ"
	barzoy888                           = "UCIYF1orTg4nDvv29ORoKXyA"
	_7playstudiopersian                 = "UCJ3QUKcd9yOhiJzZBR2zT3Q"
	eazypurple                          = "UCJ6Hd8S39g7pe65A6k42NZw"
	GennadyM                            = "UCjZlOrmC7hi_1kMA6BKZXeQ"
	GTGujarati                          = "UCKnYZmYFdFX4hP09Hh8k8jw"
	annapavlidou                        = "UCKPSHyCFIDtMqzbJ_Qx-7aw"
	Tamilmemesclub                      = "UCL_DPW38jcSGWy0Fu0A_hCg"
	QuickReactionTeam                   = "UCl--LVcANGKHJYN7O8uY6OA"
	rodrigojoseoficial                  = "UClMsSb57Jvexl1a0a2c2_CQ"
	AzadChaiwala                        = "UCM5QNdoIefx6eumjPk8ZTMw"
	lafenixextras                       = "UCM7zougvyPtfd9zw5HmAGlg"
	JustSawAGhost                       = "UCmQYXE03n6qlF9FhtAPztcw"
	mirchibangla                        = "UCmzj6hXrPZ_AwIZ8lgo-HuQ"
	marsaguirre                         = "UCN3OMvdU7ySvE6meGjo9omg"
	JustGoool                           = "UCn7kTV_syU_rI0jANxrpaFg"
	nossocanal                          = "UCNDb9jdx5Jz-C7hWJv3y6Fw"
	CSNN                                = "UCnfn-8PJbjYFAJ1fqzMdcew"
	pedronedved                         = "UCNhb2Fz3sVLZcMb_arm1cMQ"
	famoosh                             = "UCnQU2QZLVG1wxHoy_9LZucA"
	Cenoritas                           = "UCo-VSGoYy_4IjXF9u4QqTiw"
	fantabobgames                       = "UCo0U1tbk3YbqiLDhkeWOviQ"
	thesoundofsoul                      = "UCOey6R7Ktnil8RZrY_xJYiA"
	SaurabhMishraJournalist             = "UCoHVCBoSfUOlQVa80dlN60g"
	YoYoHoneySinghlive                  = "UCOZTp3nGj39-snzo4QSdduA"
	CINEGLITZ                           = "UCP_p8JkLOwPcGAnxN1iVhZA"
	lovetreta                           = "UCPi3uuGrh2mxnZJMfRhcjiw"
	canalvendimeusofa                   = "UCpk58NDdaKdX0QiiA2e79tg"
	CARIBEANVIDEOS                      = "UCPynGkfzH35RKyXRccCgzNg"
	Dimon                               = "UCpZsZZ6KEqCHeAJ_Y0gEbyQ"
	Anantvijaysoni                      = "UCQ3NeFKF2yo70xxBBUDeXXQ"
	canalilhadebarbados                 = "UCR5L_Q8Tiljy7WKEQOrGKbg"
	SotomayorTV2                        = "UCR70CjRHxQilfEUgBBYpDhg"
	Recipesarab                         = "UCRbMDfkH_bPUjQsv5dHcFQA"
	_4ak10                              = "UCsbrIjDPPXuVlWApYTnQECA"
	Top50                               = "UCSjXR8uvU4PijMM_kgyVQfw"
	SoumyajitPyne                       = "UCSJXwgF3SePNfyojYG3yt6Q"
	SneksShow                           = "UCsqZSkVccnyIxSE6pZuj13g"
	casadosaber                         = "UCtvvTFp0XANyllOdmzZr9VQ"
	Entarexyt                           = "UCuDkq6yAb5zdM2zIRG6VRDw"
	UCurU6GLM4ggcLtWWAOTzlYw            = "UCurU6GLM4ggcLtWWAOTzlYw"
	promining                           = "UCviqKxlMnZqBipTzwixkGyg"
	zapkids                             = "UCx2uGhYa9EuCNmbL72hVKGg"
	Canalokok                           = "UCXyCAcjoWz9SMLGJT8dR0pA"
	SmoothGames                         = "UCY8gvG25rOZ4hWWvcQH88PQ"
	LIGATFA                             = "UCYhiIMOlDn_HEn3zvqZCDiw"
	DeHamad                             = "UCZ5G07Vw7IaV81InBgYLCcg"
	misszizi                            = "UCE55WTFs4ekJ_aWCoNEapbQ"
	creationshub                        = "UCNfELkowZPIQ-Vegmb3pIBQ"
	AlicePandora                        = "UCScRxEtwlt082_6ThI0YJbg"
	Akito0405                           = "UCXHWx1teSYIKwTGiscYaC-Q"
	dashcambristol                      = "UCD0IC8bZI-MfIutkgPHlyWA"
	adrianbonillap                      = "UCdmY7p_kC-QN8jS7rocmCSA"
	SubhamVlogs                         = "UCXtdoCLRlnLIW3Skix0rHaQ"
	DisciplesofJesusChristJeremiahPayne = "UC_y2rVsotcQcVF7LImVGs5Q"
	ModernGalaxy                        = "UCGISiGs_RL7Z1qfs1-GD6PA"
	socofilms                           = "UCyDS9p6NWHpU9XbbbYLFLBw"
	buriedone                           = "UC_lm7xXB3adOTc0T1FwyGRQ"
	bitcoin                             = "UC-f5nPBEDyUBZz_jOPBwAfQ"
	globalrashid                        = "UC2ldcEtbR7cFYadgrnW3B6Q"
	gameofbitcoins                      = "UC2WKsYBxMwx7E7ENkND-fkg"
	CryptoInsightsBrasil                = "UC4BrnREinCBUenQZi4ZU95w"
	Crypt0                              = "UCdUSSt-IEUg2eq46rD7lu_g"
	cryptomined                         = "UCGQ3XHtsH8Q9iQr9bFbgfDA"
	altcoinbuzz                         = "UCGyqEtcGQQtXyUwvcy7Gmyg"
	crypto                              = "UCiMgF08KQ4z-Gnu8o2BLOxA"
	Crypto99                            = "UCjsrdOJCAKcuBqyyqjg1cCQ"
	TheCryptoLark                       = "UCl2oCaw8hdR_kbqyqd2klIA"
	NuggetsNews                         = "UCLo66QVfEod0nNM_GzKNxmQ"
	btckyle                             = "UCNCGCxxTT10aeTgUMHW5FfQ"
	LouisThomas                         = "UCpceefaJ9vs4RYUTsO9Y3FA"
	BitsBeTrippin                       = "UCVVWXoQfMfQVuzxcLylq9aA"
	cryptocrow                          = "UCwsRWmIL5XKqFtdytBfeX0g"
	KouSuccessLeeFX                     = "UC0YkP4Fg_d8y6JFwLj3MLdg"
	TheSchiffReport                     = "UCIjuLiLHdFxYtFmWlbTGQRQ"
	MRU                                 = "UCnkEhPBMZcEO0QGu51fDFDg"
	Vidello                             = "UCwMaWqZ6SdDpTaYOW0huELw"
	kaccreative                         = "UC02O9ICMuwrfULSa0N6SiSw"
	shecooks                            = "UChZYqTJkeYV2r7WETcytdPQ"
	vegsource                           = "UClEsPxvotpTJ1Z8eu2Y97rg"
	VeganGains                          = "UCr2eKhGzPhN5RPVk5dd5o3g"
	AwesomeKnowledge                    = "UC_pC4T8vr-caiNDYU6PYTTA"
	luckyloush                          = "UCbO0Oomf_jr20Wb4V2ap5Uw"
	MichaelLuzzi                        = "UCDwLrj4DSGrw3gFvANbl_Cw"
	NakedApe                            = "UCMOaRU-YsXVgU-WahBkZqWQ"
	ParaReact                           = "UCuJKELjsmWTlJG0X7T4ZD_A"
	dullytv                             = "UCv1J91Nhn7KsxMFaxKChT3w"
	REDONKULAS                          = "UCwd_sSDZ8EQt6SEeOO2tBRA"
	RedactedTonight                     = "UCyvaZ2RHEDrgKXz43gz7CbQ"
	TonyTornado                         = "UCZu9AV3mrCCDpK_dy1qJRAQ"
	hoodlumscrafty                      = "UC0cTVjYKgAnBrXQKcICyNmA"
	Karmakut                            = "UC2B8TOklu2rSDULqAzwn5GQ"
	Draxr                               = "UC72o3j23E1wKskBDEKlmTOw"
	BigfryTV                            = "UC7FVdUA3SxDMfl4fLBGlADg"
	CSGONews                            = "UC7L6NRLyldvxWukhOHABazQ"
	ZedGaming                           = "UC7Q15H71DmB7F1iSFE1Z8LQ"
	KamiVS                              = "UC9fh15yUcGAr8iUfQaoRRpQ"
	grabthegames                        = "UCaJFEgY6ij05Fxgn6qtcX_g"
	blueplays                           = "UCbjMsFlYb2NLpjS3uDzm9ow"
	ImNotBonkers                        = "UCCyN0G77B7wnf-1AOTI3gWQ"
	GunslingerMedia                     = "UCDjs4JoXmMzvaZyPVsvjKXw"
	LeagueVoices                        = "UCdkonRjBLzr1Adf7Jhu1bWQ"
	SirPugger                           = "UCelqWKTcCvBZP_1_1iYZAWA"
	gamesoup                            = "UCGPMrF9AN_D9BrmSmMeV3hA"
	ProHenis                            = "UCIjFoXSQ9HYbcWmRVIsH-Ow"
	Dota2Divine                         = "UCiR9IHCurqVHpC821NVcW6g"
	CrazyFoxMovies                      = "UCjewtQLpJEENPLbrCtb6YpA"
	Larry                               = "UCJVdNvvuvOnthuWVQjYff2w"
	hottake                             = "UCK24784Oqb4oYQfHL9XePNg"
	Breezy                              = "UCKWRpZpcLKriWd1am9SHf8A"
	retrorgb                            = "UCLPIbBCKVH2uKGm5C4sOkew"
	Zer0Gamer                           = "UCmII34jN4rqCIsGqWkK5JEg"
	nickatnyte                          = "UCMxYQX1zaepCgmiSmwbT39w"
	DyllonStejGaming                    = "UCngaLL0QDbsAGYj7zsB8o_Q"
	bluedrake42                         = "UCNSwcDEUfIEzYdAPscXo6ZA"
	Rerez                               = "UCoFpRCAsKfWAshvLE1bYzdw"
	Op                                  = "UCowi5kFfvGXR8NqhyE6jneQ"
	kidsgamesfun                        = "UCqjGzmb2pMWSFYm0GzKeTEA"
	GamesGlitches                       = "UCRj3Q06KOxAZWHaHScrkAOw"
	Musikage                            = "UCsej4tgCoXDgVH3J7M3NMgw"
	JeffyGaming                         = "UCTZzSNnZ43XQslejT5wFdRw"
	oniblackmage                        = "UCUEF9XL3o8dZ6hvVf8jAi8Q"
	nickatnyte2                         = "UCuoTqrobMyZj0ge8LOCkiSw"
	KazeEmanuar                         = "UCuvSqzfO_LV_QzHdmEj84SQ"
	TheLinuxGamer                       = "UCv1Kcz-CuGM6mxzL3B1_Eiw"
	GaminGHD                            = "UCW-thz5HxE-goYq8yPds1Gw"
	BHGaming                            = "UCX4N3DioqqrugFeilxTkSIw"
	Potato                              = "UCxPPTDNH85HZWxrgZ3FQBYA"
	MikeyTaylorGaming                   = "UCycXj6lRWtsSqo-bZOIZePw"
	Acituanbus                          = "UCzfx1QvKjn-BxLMLBdBGgMw"
	juggling                            = "UC2fhTIbnQlFYaFzyTcmPkXg"
	KhanAcademy                         = "UC4a-Gbdw7vOaccHmFo40b9g"
	DON                                 = "UCAYrPk70AePJZSaVLKrWdfQ"
	shogogarcia                         = "UCE3yZjxDg3iI91bcNJDFnsg"
	alphalifestyleacademy               = "UCeggEaXtJu2domMahYMD_ig"
	NileRed                             = "UCFhXFikryT4aFcLkLw2LBLA"
	veritasium                          = "UCHnyfMqiRRG1u-2MsSQLbXA"
	stevecronin                         = "UCJYawZQYwjZ76mrF_US9eNg"
	jeranism                            = "UCS_FY5mR4g22L_E9t1D_ExQ"
	MinutePhysics                       = "UCUHW94eEFW7hkUMVaZz4eDg"
	_3Blue1Brown                        = "UCYO_jab_esuFRV4b17AJtAw"
	Itsrucka                            = "UC-B2LyEZcl3avG0coKeohGQ"
	unitedtaps                          = "UC5Q8e9-uutVZmiwAcEHKuzA"
	srodalmenara                        = "UC9kc1DaOy2kSzXHZgpn9kfQ"
	ShutupAndPlay                       = "UCAwuvzhah0KUw5QNihSkEwQ"
	sanx2                               = "UCJ_waKl9kjbhfXI3LJOfHvA"
	DerickWattsAndTheSundayBlues        = "UCmZhiZq7M7d73Kbey4yna_Q"
	akirathedon                         = "UCsoiSpBvkr4Y-78Pj3recUw"
	Musicoterapia                       = "UCsoSK8K4OpdMV1tqJFwO5QA"
	daydreaming                         = "UCtbuGylbRXc42pIxWey19Dg"
	RemixHolicRecords                   = "UCUW5GjwcXgbPcfk00t-GQZA"
	EDMBot                              = "UCvmUdL2NHWlj1NRiNJPI-TQ"
	TrioTravels                         = "UCdAPAmdnkdFsH5R2Hxevucg"
	caosonnguyen                        = "UCwPeW9kFId5-VbQ2LQEjVhg"
	TheAlmonteFilms                     = "UC4C_SF5koS4Q5om50b9NMTw"
	timcast                             = "UCG749Dj4V2fKa143f8sE60Q"
	SeekingTheTruth                     = "UCHrDpTVL9S0h91u9UCPgVbA"
	davidpakman                         = "UCvixJtaXuNdMPUGdOPcY8Ag"
	mikenayna                           = "UCzk08fzh5c_BhjQa1w35wtA"
	DoorMonster                         = "UC-to_wlckb-bFDtQfUZL3Kw"
	barnacules                          = "UC1MwJy1R0nGQkXxRD9p-zTQ"
	brightsunfilms                      = "UC5k3Kc0avyDJ2nG9Kxm9JmQ"
	Onision                             = "UC5OxQNCgW88FDBxeZCnrBbg"
	top10archive                        = "UCa03bf8gAS2EtffptV-_jfA"
	thought                             = "UCb0yiUQhhLV_jpY3BayJaLA"
	MothersBasement                     = "UCBs2Y3i14e1NWQxOGliatmg"
	VlogsOfKnowledge                    = "UCDUPGR6uL5uz0_hiRAQjZMQ"
	GorTheMovieGod                      = "UCHdTVw89QU6coU1MgN-9RHA"
	iamalexoconnor                      = "UCKAQLEk1GGqnPtov9EW0huQ"
	ADedits                             = "UCsX-zRuq3ovMsgFqEQLw2Bw"
	SynthCool                           = "UCxGTHsD0pLSFlFI7M7jYmBQ"
	JordanBPeterson                     = "UCL_f53ZEJxp8TtlOkHwMV9Q"
	Sciencedocumentinhindi              = "UC9SpfUF3rm-MGep5WE6FSCA"
	NurdRage                            = "UCIgKGGJkt1MrNmhq3vRibYA"
	MINDBLASTER                         = "UC_ZMqbRu44jK-EogjYyHz8g"
	DaminousPurity                      = "UCdKRWvz50QoioZFgu6Nf9og"
	ScammerRevolts                      = "UC0uJKUXiU5T41Fzawy5H6mw"
	eevblog                             = "UC2DjFE7Xf11URZqWBigcVOQ"
	Luke                                = "UC2eYFnH61tmytImy1mTYvhA"
	AppGirl                             = "UC389S4_2Yt9cei1qNDwmBrA"
	thecryptodad                        = "UC68x_TIzqCtF69fYl2_kl3w"
	ChrisWereDigital                    = "UCAPR27YUyxmgwm3Wc2WSHLw"
	alecaddd                            = "UCbmBY_XYZqCa2G0XmFA7ZWg"
	EliTheComputerGuy                   = "UCD4EOyXKjfDUhCI6jlOZZYQ"
	archetapp                           = "UCDIBBmkZIB2hjBsk1hUImdA"
	PCPlaceNZ                           = "UCf5ZTSZAKbinY03jOwylfOg"
	NaomiSexyCyborgWu                   = "UCh_ugKacslKhsGGdXP0cRRA"
	NibiruWatcher                       = "UCi62JvN-lUn7hVL3ffofADA"
	imineblocks                         = "UCjYHcWGAjUVqU49D2JOKD3w"
	Lunduke                             = "UCkK9UDm_ZNrq_rIXCz3xCGA"
	CooLoserTech                        = "UCl97rZ2Tc7KV9lktmmHNFDQ"
	TechHD                              = "UCN3bPy04Jkp3ADRtyYvXomQ"
	MiketheScrapper                     = "UCqtlJpkH_llXS_vuDExGVvw"
	eevblog2                            = "UCr-cm90DwFJC0W3f9jBs5jA"
	thecreativeone                      = "UCTikFhzCiIXfOMS7D29dvYg"
	weekendtricks                       = "UCYtAJXx0ymGPpCndn2Gt6-w"
	GBGuns                              = "UC2VOURrALs1CwVmbGlXJOPQ"
	Matsimus                            = "UCFWjEwhX6cSAKBQ28pufG3w"
	TheLateBoyScout                     = "UCZjvj5MN3BMxPFfdEKIrvxQ"
	BravoCinematografica                = "UC2ruSXQoKMgr7JXzwO0H0KA"
	PyroNation                          = "UC4ffy3n1hE7Z8q-2KVq_91Q"
	dramatuber                          = "UC4Y8mImty3gFEG9grsR6T-w"
	BarnabasNagy                        = "UC8TRZRK1sJfxKJ0tXMmGTow"
	avery                               = "UCcfjIZLDCuSqkIlkH8i4DDg"
	Tingledove                          = "UCfI7wtV6K64gVbzjH7DOA_Q"
	crmjewelers                         = "UChpFWeF84jA5JeV3YyIo3pQ"
	NerfNerd18                          = "UCIgmaEJNqvH9bU9zGMapZGg"
	SEIJIHITO                           = "UCNqUrLE6dI8fWw_u3HQkpXA"
	anvithavlogs                        = "UCsP9pYat2DEBvnvF_iFGG_w"
	YoelRekts                           = "UCZ_BcFyhIo6GdtTSrqXXepg"
	TechFox                             = "UCIp-oTSdFO7BhAJpW2d5HMQ"
)

var channelWideTags = map[string][]string{
	JustJuggling:                        {"juggling", "circus arts", "malabares"},
	SwissExperiments:                    {"science & technology", "experiments", "switzerland"},
	TechFox:                             {"technology", "reviews"},
	misszizi:                            {"art", "pop culture"},
	creationshub:                        {"art"},
	AlicePandora:                        {"art"},
	Akito0405:                           {"art"},
	dashcambristol:                      {"automotive"},
	adrianbonillap:                      {"automotive"},
	SubhamVlogs:                         {"automotive"},
	DisciplesofJesusChristJeremiahPayne: {"beliefs"},
	ModernGalaxy:                        {"beliefs"},
	socofilms:                           {"beliefs"},
	buriedone:                           {"blockchain"},
	bitcoin:                             {"blockchain"},
	globalrashid:                        {"blockchain"},
	gameofbitcoins:                      {"blockchain"},
	CryptoInsightsBrasil:                {"blockchain"},
	Crypt0:                              {"blockchain"},
	cryptomined:                         {"blockchain"},
	altcoinbuzz:                         {"blockchain", "technology"},
	crypto:                              {"blockchain"},
	Crypto99:                            {"blockchain"},
	TheCryptoLark:                       {"blockchain", "technology"},
	NuggetsNews:                         {"blockchain", "learning"},
	btckyle:                             {"blockchain"},
	LouisThomas:                         {"blockchain"},
	BitsBeTrippin:                       {"blockchain"},
	cryptocrow:                          {"blockchain"},
	KouSuccessLeeFX:                     {"economics", "learning"},
	TheSchiffReport:                     {"economics"},
	MRU:                                 {"economics"},
	Vidello:                             {"economics", "pop culture"},
	kaccreative:                         {"food", "art"},
	shecooks:                            {"food"},
	vegsource:                           {"food"},
	VeganGains:                          {"food"},
	AwesomeKnowledge:                    {"funny"},
	luckyloush:                          {"funny"},
	MichaelLuzzi:                        {"funny", "pop culture"},
	NakedApe:                            {"funny"},
	ParaReact:                           {"funny"},
	dullytv:                             {"funny"},
	REDONKULAS:                          {"funny"},
	RedactedTonight:                     {"funny", "news"},
	TonyTornado:                         {"funny"},
	hoodlumscrafty:                      {"gaming", "pop culture"},
	Karmakut:                            {"gaming"},
	Draxr:                               {"gaming", "pop culture"},
	BigfryTV:                            {"gaming"},
	CSGONews:                            {"gaming", "funny"},
	ZedGaming:                           {"gaming"},
	KamiVS:                              {"gaming"},
	grabthegames:                        {"gaming"},
	blueplays:                           {"gaming"},
	ImNotBonkers:                        {"gaming", "funny"},
	GunslingerMedia:                     {"gaming"},
	LeagueVoices:                        {"gaming"},
	SirPugger:                           {"gaming"},
	gamesoup:                            {"gaming"},
	ProHenis:                            {"gaming"},
	Dota2Divine:                         {"gaming"},
	CrazyFoxMovies:                      {"gaming", "pop culture", "technology"},
	Larry:                               {"gaming"},
	hottake:                             {"gaming", "funny"},
	Breezy:                              {"gaming", "funny"},
	retrorgb:                            {"gaming"},
	Zer0Gamer:                           {"gaming"},
	nickatnyte:                          {"gaming"},
	DyllonStejGaming:                    {"gaming"},
	bluedrake42:                         {"gaming"},
	Rerez:                               {"gaming"},
	Op:                                  {"gaming"},
	kidsgamesfun:                        {"gaming", "pop culture"},
	GamesGlitches:                       {"gaming"},
	Musikage:                            {"gaming", "pop culture"},
	JeffyGaming:                         {"gaming", "funny"},
	oniblackmage:                        {"gaming"},
	nickatnyte2:                         {"gaming"},
	KazeEmanuar:                         {"gaming"},
	TheLinuxGamer:                       {"gaming", "technology", "linux"},
	GaminGHD:                            {"gaming", "pop culture"},
	BHGaming:                            {"gaming"},
	Potato:                              {"gaming"},
	MikeyTaylorGaming:                   {"gaming"},
	Acituanbus:                          {"gaming"},
	juggling:                            {"juggling", "circus art", "malabares"},
	KhanAcademy:                         {"learning", "science"},
	DON:                                 {"learning", "pop culture"},
	shogogarcia:                         {"learning"},
	alphalifestyleacademy:               {"learning"},
	NileRed:                             {"learning", "science"},
	veritasium:                          {"learning", "science"},
	stevecronin:                         {"learning"},
	jeranism:                            {"learning"},
	MinutePhysics:                       {"learning", "science"},
	_3Blue1Brown:                        {"learning"},
	Itsrucka:                            {"music"},
	unitedtaps:                          {"music"},
	srodalmenara:                        {"music"},
	ShutupAndPlay:                       {"music", "learning"},
	sanx2:                               {"music"},
	DerickWattsAndTheSundayBlues:        {"music", "funny"},
	akirathedon:                         {"music"},
	Musicoterapia:                       {"music"},
	daydreaming:                         {"music"},
	RemixHolicRecords:                   {"music"},
	EDMBot:                              {"music"},
	TrioTravels:                         {"nature"},
	caosonnguyen:                        {"nature"},
	TheAlmonteFilms:                     {"news"},
	timcast:                             {"news", "technology"},
	SeekingTheTruth:                     {"news"},
	davidpakman:                         {"news"},
	mikenayna:                           {"news"},
	DoorMonster:                         {"pop culture", "funny"},
	barnacules:                          {"pop culture", "gaming"},
	brightsunfilms:                      {"pop culture"},
	Onision:                             {"pop culture", "funny"},
	top10archive:                        {"pop culture"},
	thought:                             {"pop culture", "learning"},
	MothersBasement:                     {"pop culture", "gaming"},
	VlogsOfKnowledge:                    {"pop culture", "gaming"},
	GorTheMovieGod:                      {"pop culture"},
	iamalexoconnor:                      {"pop culture"},
	ADedits:                             {"pop culture"},
	SynthCool:                           {"pop culture", "funny"},
	JordanBPeterson:                     {"psychology", "postmodernism", "news"},
	Sciencedocumentinhindi:              {"science"},
	NurdRage:                            {"science", "learning"},
	MINDBLASTER:                         {"sports", "funny"},
	DaminousPurity:                      {"sports", "gaming"},
	ScammerRevolts:                      {"technology"},
	eevblog:                             {"technology", "learning"},
	Luke:                                {"technology", "funny"},
	AppGirl:                             {"technology"},
	thecryptodad:                        {"technology", "blockchain"},
	ChrisWereDigital:                    {"technology"},
	alecaddd:                            {"technology", "learning"},
	EliTheComputerGuy:                   {"technology"},
	archetapp:                           {"technology", "learning"},
	PCPlaceNZ:                           {"technology"},
	NaomiSexyCyborgWu:                   {"technology"},
	NibiruWatcher:                       {"technology", "learning"},
	imineblocks:                         {"technology", "blockchain"},
	Lunduke:                             {"technology"},
	CooLoserTech:                        {"technology"},
	TechHD:                              {"technology", "learning"},
	MiketheScrapper:                     {"technology"},
	eevblog2:                            {"technology"},
	thecreativeone:                      {"technology", "gaming"},
	weekendtricks:                       {"technology"},
	GBGuns:                              {"weapons"},
	Matsimus:                            {"weapons", "gaming"},
	TheLateBoyScout:                     {"weapons"},
}
var tagsToSkip = map[string]*struct{}{
	"#hangoutsonair":                         nil,
	"#hoa":                                   nil,
	"1080p":                                  nil,
	"2":                                      nil,
	"2012":                                   nil,
	"2013":                                   nil,
	"2014":                                   nil,
	"2015":                                   nil,
	"2016":                                   nil,
	"2017":                                   nil,
	"2018":                                   nil,
	"2019":                                   nil,
	"360":                                    nil,
	"3d":                                     nil,
	"60fps":                                  nil,
	"720p":                                   nil,
	"achievement":                            nil,
	"action":                                 nil,
	"adam":                                   nil,
	"addon":                                  nil,
	"adityanath":                             nil,
	"africa":                                 nil,
	"african american":                       nil,
	"akshay kumar":                           nil,
	"alien":                                  nil,
	"all":                                    nil,
	"alpha":                                  nil,
	"amazing":                                nil,
	"america":                                nil,
	"amerika":                                nil,
	"anal":                                   nil,
	"and":                                    nil,
	"asia":                                   nil,
	"ass":                                    nil,
	"atlanta":                                nil,
	"atmospheric":                            nil,
	"attack":                                 nil,
	"aughad":                                 nil,
	"auto imagen":                            nil,
	"aventure":                               nil,
	"awesome":                                nil,
	"baba":                                   nil,
	"babas":                                  nil,
	"bakri":                                  nil,
	"bandar":                                 nil,
	"base":                                   nil,
	"battle":                                 nil,
	"battlefield":                            nil,
	"beard":                                  nil,
	"best":                                   nil,
	"beta":                                   nil,
	"betv":                                   nil,
	"bhagwa":                                 nil,
	"bharat":                                 nil,
	"bhawreshwara":                           nil,
	"bhoj":                                   nil,
	"big boobs":                              nil,
	"big dick":                               nil,
	"bill still":                             nil,
	"black":                                  nil,
	"blackpeace72":                           nil,
	"blog":                                   nil,
	"blowjob":                                nil,
	"blue":                                   nil,
	"bob lennon":                             nil,
	"bob":                                    nil,
	"bollywood news":                         nil,
	"bollywood tashan":                       nil,
	"bollywood":                              nil,
	"boobs":                                  nil,
	"bounty":                                 nil,
	"build":                                  nil,
	"call":                                   nil,
	"camera":                                 nil,
	"campaign":                               nil,
	"canada":                                 nil,
	"challenge":                              nil,
	"challenges":                             nil,
	"champion":                               nil,
	"channel":                                nil,
	"chiara ferragni":                        nil,
	"chickens":                               nil,
	"china":                                  nil,
	"city":                                   nil,
	"clan":                                   nil,
	"clans":                                  nil,
	"clash of clans":                         nil,
	"clash royale":                           nil,
	"clash":                                  nil,
	"classic":                                nil,
	"colorful":                               nil,
	"commentary":                             nil,
	"compilation":                            nil,
	"convention":                             nil,
	"cool":                                   nil,
	"coplanet":                               nil,
	"cow":                                    nil,
	"craft":                                  nil,
	"crazy":                                  nil,
	"creampie":                               nil,
	"creative beard":                         nil,
	"csgo":                                   nil,
	"cumshot":                                nil,
	"custom":                                 nil,
	"cyber locks":                            nil,
	"daily celebration":                      nil,
	"daily holidays":                         nil,
	"dark":                                   nil,
	"dave pacman":                            nil,
	"dave":                                   nil,
	"david di franco":                        nil,
	"david packman":                          nil,
	"david pacman":                           nil,
	"david pakman show":                      nil,
	"david pakman":                           nil,
	"david":                                  nil,
	"davidpakman.com":                        nil,
	"de":                                     nil,
	"dead":                                   nil,
	"dean":                                   nil,
	"death noise":                            nil,
	"death voice":                            nil,
	"deep":                                   nil,
	"defense":                                nil,
	"demo":                                   nil,
	"depression":                             nil,
	"deutsch":                                nil,
	"dfx":                                    nil,
	"dharm":                                  nil,
	"difranco":                               nil,
	"direct":                                 nil,
	"dnb portal":                             nil,
	"dnb":                                    nil,
	"dnbportal":                              nil,
	"download":                               nil,
	"drive":                                  nil,
	"easy":                                   nil,
	"eatmydiction1":                          nil,
	"eeuu":                                   nil,
	"empire":                                 nil,
	"ending":                                 nil,
	"energy":                                 nil,
	"english":                                nil,
	"entertainment":                          nil,
	"episode":                                nil,
	"erik":                                   nil,
	"europe":                                 nil,
	"fails":                                  nil,
	"fanta":                                  nil,
	"fantabobgames":                          nil,
	"farm":                                   nil,
	"farming":                                nil,
	"fast":                                   nil,
	"festival":                               nil,
	"fight":                                  nil,
	"fighter":                                nil,
	"fighting":                               nil,
	"fights":                                 nil,
	"fireworks":                              nil,
	"first":                                  nil,
	"fist":                                   nil,
	"florida":                                nil,
	"footage":                                nil,
	"for":                                    nil,
	"foto":                                   nil,
	"fr":                                     nil,
	"français":                               nil,
	"free":                                   nil,
	"friends":                                nil,
	"fuck":                                   nil,
	"full hd":                                nil,
	"full":                                   nil,
	"fun":                                    nil,
	"futuristic":                             nil,
	"gaay":                                   nil,
	"gameplay fr":                            nil,
	"gamerworf":                              nil,
	"gamingoncaffeine":                       nil,
	"garena":                                 nil,
	"gay":                                    nil,
	"george senda":                           nil,
	"german":                                 nil,
	"get":                                    nil,
	"gift":                                   nil,
	"girl":                                   nil,
	"girls":                                  nil,
	"giveaway":                               nil,
	"glitch":                                 nil,
	"good":                                   nil,
	"google":                                 nil,
	"gopro":                                  nil,
	"gorthemoviegod":                         nil,
	"gps":                                    nil,
	"great":                                  nil,
	"green":                                  nil,
	"gt":                                     nil,
	"guide":                                  nil,
	"guy":                                    nil,
	"handjob":                                nil,
	"hangouts on air":                        nil,
	"hard fucking":                           nil,
	"hcg":                                    nil,
	"hd":                                     nil,
	"hdtv":                                   nil,
	"hentai":                                 nil,
	"heroes of newerth":                      nil,
	"heroes":                                 nil,
	"high":                                   nil,
	"highlights":                             nil,
	"holiday everyday":                       nil,
	"hot":                                    nil,
	"house flipper":                          nil,
	"house party":                            nil,
	"houseparty":                             nil,
	"hungarian vlog":                         nil,
	"imagen":                                 nil,
	"imovie":                                 nil,
	"in":                                     nil,
	"inc":                                    nil,
	"india":                                  nil,
	"indonesia":                              nil,
	"industry (organization sector)":         nil,
	"influencer":                             nil,
	"injured":                                nil,
	"instagram":                              nil,
	"interior":                               nil,
	"interview":                              nil,
	"intro":                                  nil,
	"is":                                     nil,
	"it":                                     nil,
	"ita":                                    nil,
	"jaanwar":                                nil,
	"japan":                                  nil,
	"jay's":                                  nil,
	"jeux vidéo":                             nil,
	"jeux":                                   nil,
	"jew":                                    nil,
	"jnrsnr":                                 nil,
	"jnrsnrgaming":                           nil,
	"joe":                                    nil,
	"john sonmez":                            nil,
	"johnsp69":                               nil,
	"jump":                                   nil,
	"junior senior gaming":                   nil,
	"junior senior":                          nil,
	"kag3 entertainment":                     nil,
	"kag3":                                   nil,
	"karmakut":                               nil,
	"katrina kaif":                           nil,
	"kevin":                                  nil,
	"kids":                                   nil,
	"king":                                   nil,
	"kokesh":                                 nil,
	"kristomaster4":                          nil,
	"kutta":                                  nil,
	"la":                                     nil,
	"lance scurvin":                          nil,
	"lancescurv":                             nil,
	"latest bollywood news":                  nil,
	"launch":                                 nil,
	"legends":                                nil,
	"lennon":                                 nil,
	"liberal news":                           nil,
	"life is strange":                        nil,
	"life":                                   nil,
	"like":                                   nil,
	"liquid":                                 nil,
	"live stream":                            nil,
	"live":                                   nil,
	"livestream":                             nil,
	"london":                                 nil,
	"lp":                                     nil,
	"magyar vlog":                            nil,
	"magyar vlogger":                         nil,
	"make":                                   nil,
	"man":                                    nil,
	"map":                                    nil,
	"martinez ca":                            nil,
	"maskedmage":                             nil,
	"mature":                                 nil,
	"michigan":                               nil,
	"mine":                                   nil,
	"minimal":                                nil,
	"mission":                                nil,
	"mmr":                                    nil,
	"mobile":                                 nil,
	"mode":                                   nil,
	"modi":                                   nil,
	"moments":                                nil,
	"monster":                                nil,
	"montage":                                nil,
	"moon":                                   nil,
	"mortal":                                 nil,
	"multicolored":                           nil,
	"music":                                  nil,
	"my":                                     nil,
	"narendra":                               nil,
	"navidad":                                nil,
	"neurofunk":                              nil,
	"new":                                    nil,
	"nickatnyte":                             nil,
	"nidge":                                  nil,
	"night":                                  nil,
	"nma":                                    nil,
	"no commentary":                          nil,
	"no":                                     nil,
	"noise":                                  nil,
	"north":                                  nil,
	"nsfw":                                   nil,
	"obiettivo":                              nil,
	"of":                                     nil,
	"official":                               nil,
	"old":                                    nil,
	"on":                                     nil,
	"one":                                    nil,
	"opening":                                nil,
	"ops":                                    nil,
	"orange":                                 nil,
	"outrageous":                             nil,
	"overview":                               nil,
	"packman":                                nil,
	"pakman":                                 nil,
	"part 1":                                 nil,
	"part":                                   nil,
	"party":                                  nil,
	"paul":                                   nil,
	"ped":                                    nil,
	"pewdiepie":                              nil,
	"pig":                                    nil,
	"pittsburgh pa":                          nil,
	"plus":                                   nil,
	"podcastradio":                           nil,
	"police":                                 nil,
	"porn":                                   nil,
	"porno":                                  nil,
	"power":                                  nil,
	"pradesh":                                nil,
	"prakriti":                               nil,
	"premiere":                               nil,
	"preview":                                nil,
	"price":                                  nil,
	"productions":                            nil,
	"progressive news":                       nil,
	"progressive podcast":                    nil,
	"pussy":                                  nil,
	"quality":                                nil,
	"radio":                                  nil,
	"raebareli":                              nil,
	"rage":                                   nil,
	"raid":                                   nil,
	"rants":                                  nil,
	"react":                                  nil,
	"reaction":                               nil,
	"real":                                   nil,
	"red":                                    nil,
	"relationships":                          nil,
	"release":                                nil,
	"replay":                                 nil,
	"replica":                                nil,
	"review":                                 nil,
	"road":                                   nil,
	"russia":                                 nil,
	"sadhguru":                               nil,
	"sadhu":                                  nil,
	"samaj":                                  nil,
	"sant":                                   nil,
	"santa":                                  nil,
	"scary":                                  nil,
	"scene":                                  nil,
	"scoope":                                 nil,
	"scurv":                                  nil,
	"scurvin":                                nil,
	"segui":                                  nil,
	"series":                                 nil,
	"sex":                                    nil,
	"sexy":                                   nil,
	"shahrukh khan":                          nil,
	"sharefactory™":                          nil,
	"shield":                                 nil,
	"shooter":                                nil,
	"show":                                   nil,
	"shyam":                                  nil,
	"sikh":                                   nil,
	"silver":                                 nil,
	"simple programmer":                      nil,
	"simpleprogrammer.com":                   nil,
	"slime":                                  nil,
	"solo":                                   nil,
	"sonny daniel vlogs":                     nil,
	"sonny daniel":                           nil,
	"sonnydaniel":                            nil,
	"source":                                 nil,
	"speed":                                  nil,
	"spreaker":                               nil,
	"squad ops":                              nil,
	"squad":                                  nil,
	"sri":                                    nil,
	"states":                                 nil,
	"story":                                  nil,
	"street":                                 nil,
	"suar":                                   nil,
	"super":                                  nil,
	"support":                                nil,
	"szekely":                                nil,
	"szekelyvegan":                           nil,
	"székely vegán":                          nil,
	"székelyvegán":                           nil,
	"taiwanese animation":                    nil,
	"taiwanese animators":                    nil,
	"talk radio":                             nil,
	"talk":                                   nil,
	"tanyázás":                               nil,
	"tdps":                                   nil,
	"team":                                   nil,
	"television":                             nil,
	"terror":                                 nil,
	"test":                                   nil,
	"texas":                                  nil,
	"thailand":                               nil,
	"the brotherhood of gaming":              nil,
	"the david pakman show":                  nil,
	"the guy from pittsburgh":                nil,
	"the kag3 gaming":                        nil,
	"the kag3":                               nil,
	"the lancescurv show":                    nil,
	"the":                                    nil,
	"thecreativeone":                         nil,
	"thefantasio974":                         nil,
	"time":                                   nil,
	"tips":                                   nil,
	"to":                                     nil,
	"tom":                                    nil,
	"tomo news":                              nil,
	"tomonews":                               nil,
	"tona":                                   nil,
	"top":                                    nil,
	"total":                                  nil,
	"totka":                                  nil,
	"trevor":                                 nil,
	"trick":                                  nil,
	"tricks":                                 nil,
	"trofeo":                                 nil,
	"trolling":                               nil,
	"true":                                   nil,
	"truetotalempireinc":                     nil,
	"turbo":                                  nil,
	"tv":                                     nil,
	"uct-wqktykk1_70u4bb4k4lq":               nil,
	"uk":                                     nil,
	"ultimate":                               nil,
	"uniqornaments":                          nil,
	"unique":                                 nil,
	"united":                                 nil,
	"until dawn":                             nil,
	"up":                                     nil,
	"update":                                 nil,
	"us":                                     nil,
	"usa":                                    nil,
	"uttar":                                  nil,
	"vaanar":                                 nil,
	"vagina":                                 nil,
	"video":                                  nil,
	"videos":                                 nil,
	"vlog":                                   nil,
	"vlogger":                                nil,
	"vlogs":                                  nil,
	"voice over":                             nil,
	"voice":                                  nil,
	"vs":                                     nil,
	"vulcanhdgaming":                         nil,
	"waale":                                  nil,
	"white":                                  nil,
	"willie pelissier":                       nil,
	"willie":                                 nil,
	"win":                                    nil,
	"with":                                   nil,
	"women":                                  nil,
	"wordofgod":                              nil,
	"wounded":                                nil,
	"wow":                                    nil,
	"x320":                                   nil,
	"xxx":                                    nil,
	"you":                                    nil,
	"youtube capture":                        nil,
	"youtube editor":                         nil,
	"youtube":                                nil,
	"youtuber":                               nil,
	"ytquality=high":                         nil,
	"{5859dfec-026f-46ba-bea0-02bf43aa1a6f}": nil,
	"игра":                                   nil,
	"игры для девочек":                       nil,
	"игры для мальчиков":                     nil,
	"игры":                                   nil,
	"летсплей":                               nil,
	"прохождение игры":                       nil,
	"прохождение":                            nil,
	"рпг":                                    nil,
	"เกม":                                    nil,
}
var mapAndReplace = map[string]string{
	"nfl":                            Sports,
	"minecraft (award-winning work)": Gaming,
	"minecraft survival":             Gaming,
	"modded minecraft":               Gaming,
	"minecraft movie":                Gaming,
	"minecraft videos":               Gaming,
}

var mapAndKeep = map[string]string{
	"dance":                               Art,
	"design":                              Art,
	"fashion":                             Art,
	"creative hairstyles":                 Art,
	"colorful hair":                       Art,
	"cool long hair":                      Art,
	"cartoon":                             Art,
	"comic":                               Art,
	"comics":                              Art,
	"beauty":                              Art,
	"car":                                 Automotive,
	"cars":                                Automotive,
	"4x4":                                 Automotive,
	"automobile":                          Automotive,
	"autos":                               Automotive,
	"carros":                              Automotive,
	"suv":                                 Automotive,
	"truck":                               Automotive,
	"garage":                              Automotive,
	"auto":                                Automotive,
	"crash":                               Automotive,
	"driving":                             Automotive,
	"race":                                Automotive,
	"bmw":                                 Automotive,
	"vehicle":                             Automotive,
	"benz":                                Automotive,
	"auto show (event)":                   Automotive,
	"engine":                              Automotive,
	"mercedes":                            Automotive,
	"motorcycle":                          Automotive,
	"porsche":                             Automotive,
	"bus":                                 Automotive,
	"jeep":                                Automotive,
	"secular talk":                        Beliefs,
	"atheist":                             Beliefs,
	"atheism":                             Beliefs,
	"agnostic":                            Beliefs,
	"freedom":                             Beliefs,
	"secular":                             Beliefs,
	"religion":                            Beliefs,
	"liberty":                             Beliefs,
	"christian":                           Beliefs,
	"god":                                 Beliefs,
	"libertarian":                         Beliefs,
	"guru":                                Beliefs,
	"muslim":                              Beliefs,
	"voluntarism":                         Beliefs,
	"yogi":                                Beliefs,
	"mystic":                              Beliefs,
	"bible":                               Beliefs,
	"jesus":                               Beliefs,
	"anarchism":                           Beliefs,
	"christianity":                        Beliefs,
	"anarchy":                             Beliefs,
	"voluntaryist":                        Beliefs,
	"activism":                            Beliefs,
	"hindu":                               Beliefs,
	"peace":                               Beliefs,
	"love":                                Beliefs,
	"death":                               Beliefs,
	"christmas":                           Beliefs,
	"lord":                                Beliefs,
	"bitcoin":                             Blockchain,
	"cryptocurrency":                      Blockchain,
	"crypto":                              Blockchain,
	"ethereum":                            Blockchain,
	"btc":                                 Blockchain,
	"airdrop":                             Blockchain,
	"litecoin":                            Blockchain,
	"eth":                                 Blockchain,
	"mining":                              Blockchain,
	"bitcoin news":                        Blockchain,
	"ico":                                 Blockchain,
	"crypto news":                         Blockchain,
	"token":                               Blockchain,
	"token free":                          Blockchain,
	"bitcoin price":                       Blockchain,
	"altcoins":                            Blockchain,
	"cryptocurrency news":                 Blockchain,
	"altcoin":                             Blockchain,
	"coin":                                Blockchain,
	"ripple":                              Blockchain,
	"free bitcoin":                        Blockchain,
	"ltc":                                 Blockchain,
	"dash":                                Blockchain,
	"eos":                                 Blockchain,
	"steemit":                             Blockchain,
	"coinbase":                            Blockchain,
	"cryptocurrencies":                    Blockchain,
	"money":                               Economics,
	"economy":                             Economics,
	"gold":                                Economics,
	"federal reserve":                     Economics,
	"market":                              Economics,
	"investing":                           Economics,
	"dollar":                              Economics,
	"monetary reform":                     Economics,
	"economic":                            Economics,
	"recession":                           Economics,
	"trading":                             Economics,
	"finance":                             Economics,
	"currency":                            Economics,
	"business":                            Economics,
	"health":                              Food,
	"vegan":                               Food,
	"fruits":                              Food,
	"vegán":                               Food,
	"food holidays":                       Food,
	"fun":                                 Funny,
	"funny":                               Funny,
	"hilarious":                           Funny,
	"lol":                                 Funny,
	"humor":                               Funny,
	"funny moments":                       Funny,
	"comedy":                              Funny,
	"parody":                              Funny,
	"funny video":                         Funny,
	"silly":                               Funny,
	"humour":                              Funny,
	"satire":                              Funny,
	"walkthrough":                         Gaming,
	"#ps4live":                            Gaming,
	"twitch":                              Gaming,
	"#ps4share":                           Gaming,
	"gaming":                              Gaming,
	"video game":                          Gaming,
	"playstation":                         Gaming,
	"xbox one":                            Gaming,
	"video games":                         Gaming,
	"lets":                                Gaming,
	"nintendo":                            Gaming,
	"fps":                                 Gaming,
	"video game (industry)":               Gaming,
	"multiplayer":                         Gaming,
	"fortnite":                            Gaming,
	"dota 2":                              Gaming,
	"xbox 360":                            Gaming,
	"sony computer entertainment":         Gaming,
	"pc games":                            Gaming,
	"call of duty":                        Gaming,
	"pokemon":                             Gaming,
	"league":                              Gaming,
	"let's play":                          Gaming,
	"let's":                               Gaming,
	"league of legends":                   Gaming,
	"league of legends (video game)":      Gaming,
	"cod":                                 Gaming,
	"pubg":                                Gaming,
	"xbox":                                Gaming,
	"pvp":                                 Gaming,
	"videogame":                           Gaming,
	"dota":                                Gaming,
	"games":                               Gaming,
	"let's play fr":                       Gaming,
	"mario":                               Gaming,
	"pc game":                             Gaming,
	"ps4":                                 Gaming,
	"let":                                 Gaming,
	"fallout 4":                           Gaming,
	"shooter game (media genre)":          Gaming,
	"playthrough":                         Gaming,
	"gamer":                               Gaming,
	"wii":                                 Gaming,
	"playing":                             Gaming,
	"computer game games":                 Gaming,
	"gameplay":                            Gaming,
	"arcade":                              Gaming,
	"video game culture":                  Gaming,
	"xbox 360 (video game platform)":      Gaming,
	"rpg":                                 Gaming,
	"game":                                Gaming,
	"gta":                                 Gaming,
	"let’s play":                          Gaming,
	"minecraft":                           Gaming,
	"skyrim":                              Gaming,
	"action-adventure game (media genre)": Gaming,
	"sony interactive entertainment":      Gaming,
	"dlc":                                 Gaming,
	"xbox360":                             Gaming,
	"zelda":                               Gaming,
	"steam":                               Gaming,
	"capcom":                              Gaming,
	"fortnite battle royale":              Gaming,
	"singleplayer":                        Gaming,
	"pacman":                              Gaming,
	"gta 5":                               Gaming,
	"grand theft auto v":                  Gaming,
	"valve":                               Gaming,
	"massively multiplayer online role-playing game (video game genre)": Gaming,
	"#gaming":                               Gaming,
	"halo":                                  Gaming,
	"computer games":                        Gaming,
	"nintendo switch":                       Gaming,
	"3d games":                              Gaming,
	"ps2":                                   Gaming,
	"role-playing video game (media genre)": Gaming,
	"supercell":                             Gaming,
	"dota2":                                 Gaming,
	"sims":                                  Gaming,
	"pc gaming":                             Gaming,
	"playstation 4":                         Gaming,
	"battle royale":                         Gaming,
	"rpg games":                             Gaming,
	"sims 4":                                Gaming,
	"roblox":                                Gaming,
	"mw3":                                   Gaming,
	"far cry 5":                             Gaming,
	"squad gameplay":                        Gaming,
	"tf2":                                   Gaming,
	"ps3":                                   Gaming,
	"resident evil":                         Gaming,
	"indie game":                            Gaming,
	"overwatch":                             Gaming,
	"brawl":                                 Gaming,
	"call of duty®: black ops iii":          Gaming,
	"console":                               Gaming,
	"mugen":                                 Gaming,
	"play":                                  Gaming,
	"mass effect":                           Gaming,
	"ubisoft":                               Gaming,
	"minecraft: playstation®4 edition":      Gaming,
	"rts":                                   Gaming,
	"mmorpg":                                Gaming,
	"first person shooter":                  Gaming,
	"action role-playing game (video game genre)": Gaming,
	"role-playing game (game genre)":              Gaming,
	"sonic":                                       Gaming,
	"videogames":                                  Gaming,
	"level":                                       Gaming,
	"sims 3":                                      Gaming,
	"esports":                                     Gaming,
	"minecraft (video game)":                      Gaming,
	"simulator":                                   Gaming,
	"horrible gamers":                             Gaming,
	"moba":                                        Gaming,
	"the game":                                    Gaming,
	"link":                                        Gaming,
	"game reviews":                                Gaming,
	"sims 5":                                      Gaming,
	"snk":                                         Gaming,
	"online games":                                Gaming,
	"squad game":                                  Gaming,
	"switch":                                      Gaming,
	"lets play":                                   Gaming,
	"portal":                                      Gaming,
	"boss":                                        Gaming,
	"mod":                                         Gaming,
	"strategy":                                    Gaming,
	"stream":                                      Gaming,
	"galaxy":                                      Gaming,
	"tutorial":                                    Learning,
	"how to":                                      Learning,
	"educational":                                 Learning,
	"how":                                         Learning,
	"salman khan":                                 Learning,
	"education":                                   Learning,
	"help":                                        Learning,
	"lessons":                                     Learning,
	"online learning":                             Learning,
	"diy":                                         Learning,
	"school":                                      Learning,
	"how-to":                                      Learning,
	"howto":                                       Learning,
	"advice":                                      Learning,
	"history":                                     Learning,
	"inspirational":                               Learning,
	"drum and bass":                               Music,
	"bass":                                        Music,
	"dubstep":                                     Music,
	"song":                                        Music,
	"rap":                                         Music,
	"remix":                                       Music,
	"drum":                                        Music,
	"house":                                       Music,
	"drum bass":                                   Music,
	"darkstep":                                    Music,
	"guitar":                                      Music,
	"drumstep":                                    Music,
	"techstep":                                    Music,
	"metal":                                       Music,
	"dj":                                          Music,
	"rock":                                        Music,
	"new music":                                   Music,
	"hip hop":                                     Music,
	"cover":                                       Music,
	"instrumental":                                Music,
	"electronic":                                  Music,
	"alternative":                                 Music,
	"space":                                       Nature,
	"plants":                                      Nature,
	"trees":                                       Nature,
	"eggs":                                        Nature,
	"flowers":                                     Nature,
	"beach":                                       Nature,
	"rainbow":                                     Nature,
	"water":                                       Nature,
	"animals":                                     Nature,
	"floral":                                      Nature,
	"gardening":                                   Nature,
	"garden":                                      Nature,
	"world":                                       Nature,
	"dog":                                         Nature,
	"travel":                                      Nature,
	"rv":                                          Nature,
	"survival":                                    Nature,
	"outdoor":                                     Nature,
	"full time rving":                             Nature,
	"rv park":                                     Nature,
	"travel trailer":                              Nature,
	"environment":                                 Nature,
	"news":                                        News,
	"democrat":                                    News,
	"liberal":                                     News,
	"progressive":                                 News,
	"government":                                  News,
	"republican":                                  News,
	"conservative":                                News,
	"commentary":                                  News,
	"political":                                   News,
	"cnn":                                         News,
	"cbs":                                         News,
	"msnbc":                                       News,
	"fox news":                                    News,
	"fox":                                         News,
	"nbc":                                         News,
	"politics":                                    News,
	"news & politics":                             News,
	"senate":                                      News,
	"congress":                                    News,
	"house of representatives":                    News,
	"video news":                                  News,
	"animated news":                               News,
	"como":                                        News,
	"putin":                                       News,
	"news radio":                                  News,
	"hillary clinton":                             News,
	"progressive talk":                            News,
	"trump":                                       News,
	"barack obama":                                News,
	"donald trump":                                News,
	"newscast":                                    News,
	"racism":                                      News,
	"war":                                         News,
	"horror":                                      PopCulture,
	"zombies":                                     PopCulture,
	"next media animation":                        PopCulture,
	"adventure":                                   PopCulture,
	"epic":                                        PopCulture,
	"reviews":                                     PopCulture,
	"movie":                                       PopCulture,
	"indie":                                       PopCulture,
	"pro":                                         PopCulture,
	"animation":                                   PopCulture,
	"zombie":                                      PopCulture,
	"star":                                        PopCulture,
	"film":                                        PopCulture,
	"magic":                                       PopCulture,
	"ninja":                                       PopCulture,
	"top 10":                                      PopCulture,
	"wwe":                                         PopCulture,
	"loot":                                        PopCulture,
	"marvel":                                      PopCulture,
	"manga":                                       PopCulture,
	"puzzle":                                      PopCulture,
	"hero":                                        PopCulture,
	"retro":                                       PopCulture,
	"disney":                                      PopCulture,
	"dc":                                          PopCulture,
	"next animation studio":                       PopCulture,
	"cosplay":                                     PopCulture,
	"dragoncon":                                   PopCulture,
	"comiccon":                                    PopCulture,
	"stories":                                     PopCulture,
	"unboxing":                                    PopCulture,
	"fail":                                        PopCulture,
	"dragon":                                      PopCulture,
	"meme":                                        PopCulture,
	"family friendly":                             PopCulture,
	"random":                                      PopCulture,
	"vlogging":                                    PopCulture,
	"react":                                       PopCulture,
	"trailer":                                     PopCulture,
	"reaction":                                    PopCulture,
	"toys":                                        PopCulture,
	"reacts":                                      PopCulture,
	"cute":                                        PopCulture,
	"weed":                                        PopCulture,
	"drunkfx":                                     PopCulture,
	"happy":                                       PopCulture,
	"anime":                                       PopCulture,
	"family":                                      PopCulture,
	"memes":                                       PopCulture,
	"wtf":                                         PopCulture,
	"racing":                                      Sports,
	"sport":                                       Sports,
	"training":                                    Sports,
	"football":                                    Sports,
	"computer":                                    Technology,
	"online":                                      Technology,
	"mods":                                        Technology,
	"tech":                                        Technology,
	"linux":                                       Technology,
	"sony":                                        Technology,
	"iphone":                                      Technology,
	"pc":                                          Technology,
	"ram":                                         Technology,
	"samsung":                                     Technology,
	"software":                                    Technology,
	"windows":                                     Technology,
	"programming":                                 Technology,
	"simulation":                                  Technology,
	"facebook":                                    Technology,
	"desktop":                                     Technology,
	"install":                                     Technology,
	"microsoft":                                   Technology,
	"twitter":                                     Technology,
	"android":                                     Technology,
	"mobile":                                      Technology,
	"apple":                                       Technology,
	"ios":                                         Technology,
	"server":                                      Technology,
	"nvidia":                                      Technology,
	"open source":                                 Technology,
	"apps":                                        Technology,
	"laptop":                                      Technology,
	"podcast":                                     Technology,
	"hack":                                        Technology,
	"guns":                                        Weapons,
	"shooting":                                    Weapons,
	"trophy":                                      Weapons,
	"combat":                                      Weapons,
	"fire":                                        Weapons,
	"gun":                                         Weapons,
	"tactical":                                    Weapons,
	"firearms":                                    Weapons,
	"rocket":                                      Weapons,
	"sniper":                                      Weapons,
	"knife":                                       Weapons,
	"trap":                                        Weapons,
	"assault":                                     Weapons,
}
