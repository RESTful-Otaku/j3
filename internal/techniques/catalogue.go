// Package techniques provides a deterministic, searchable catalogue of
// Japanese Jiu Jitsu techniques for the officials console and display views.
//
// The catalogue is static data: point values are defaults only and may be
// overridden by an event's configuration. Nothing in this package depends on
// the database, so the same list is available offline on a venue LAN box.
package techniques

import (
	"strings"
	"sync"
)

// Technique is a single catalogue entry. Kind is "score", "penalty" or
// "submission"; Card is only meaningful for penalty entries and uses
// "yellow", "red", "shido" or "hansoku".
type Technique struct {
	ID         string   `json:"id"`
	English    string   `json:"english"`
	Romaji     string   `json:"romaji"`
	Japanese   string   `json:"japanese,omitempty"`
	Group      string   `json:"group"`
	PointKey   string   `json:"pointKey"`
	Kind       string   `json:"kind"`           // "score", "penalty", or "submission"
	Card       string   `json:"card,omitempty"` // "yellow", "red", "shido", "hansoku" for penalties
	Points     int      `json:"points"`         // default points when no event config override
	Tags       []string `json:"tags"`
	InstantWin bool     `json:"instantWin"`
}

// Group is an ordered catalogue section. PointKey is the default scoring
// bucket used when an event has no override for the group.
type Group struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	PointKey string `json:"pointKey"`
}

// Stable group identifiers, used by the catalogue and by callers filtering
// the console palette.
const (
	GroupStrikes          = "strikes"
	GroupKicks            = "kicks"
	GroupElbows           = "elbows"
	GroupKnees            = "knees"
	GroupPalmStrikes      = "palm_strikes"
	GroupNerveStrikes     = "nerve_strikes"
	GroupThrows           = "throws"
	GroupTakedowns        = "takedowns"
	GroupSweeps           = "sweeps"
	GroupReaps            = "reaps"
	GroupHooks            = "hooks"
	GroupSacrifice        = "sacrifice_techniques"
	GroupGroundFighting   = "ground_fighting"
	GroupHoldDowns        = "hold_downs"
	GroupPins             = "pins"
	GroupSubmissions      = "submissions"
	GroupChokes           = "chokes"
	GroupJointLocks       = "joint_locks"
	GroupControls         = "controls"
	GroupBreaks           = "breaks"
	GroupStandingVariants = "standing_variants"
	GroupGroundVariants   = "ground_variants"
	GroupKata             = "kata"
	GroupPenalties        = "penalties"
)

// groupList is the canonical display order of the catalogue.
var groupList = []Group{
	{GroupStrikes, "Strikes", "strike"},
	{GroupKicks, "Kicks", "kick"},
	{GroupElbows, "Elbows", "strike"},
	{GroupKnees, "Knees", "strike"},
	{GroupPalmStrikes, "Palm Strikes", "strike"},
	{GroupNerveStrikes, "Nerve Strikes (Kyusho)", "strike"},
	{GroupThrows, "Throws", "throw"},
	{GroupTakedowns, "Takedowns", "takedown"},
	{GroupSweeps, "Sweeps", "sweep"},
	{GroupReaps, "Reaps", "reap"},
	{GroupHooks, "Hooks", "hook"},
	{GroupSacrifice, "Sacrifice Techniques", "throw"},
	{GroupGroundFighting, "Ground Fighting", "ground"},
	{GroupHoldDowns, "Hold Downs", "hold"},
	{GroupPins, "Pins", "hold"},
	{GroupSubmissions, "Submissions", "submission"},
	{GroupChokes, "Chokes", "submission"},
	{GroupJointLocks, "Joint Locks", "submission"},
	{GroupControls, "Controls", "control"},
	{GroupBreaks, "Breaks & Escapes", "break"},
	{GroupStandingVariants, "Standing Variants", "standing"},
	{GroupGroundVariants, "Ground Variants", "ground"},
	{GroupKata, "Kata", "kata"},
	{GroupPenalties, "Penalties", "penalty"},
}

// catalogue is built once in source order; no map iteration is ever used to
// produce public results.
var catalogue = buildCatalogue()

func buildCatalogue() []Technique {
	var out []Technique
	add := func(id, english, romaji, japanese, group, pointKey, kind, card string, points int, tags []string, instantWin bool) {
		out = append(out, Technique{
			ID:         id,
			English:    english,
			Romaji:     romaji,
			Japanese:   japanese,
			Group:      group,
			PointKey:   pointKey,
			Kind:       kind,
			Card:       card,
			Points:     points,
			Tags:       append([]string(nil), tags...),
			InstantWin: instantWin,
		})
	}

	// Strikes ---------------------------------------------------------------
	add("jab", "Jab", "Jab", "", GroupStrikes, "strike", "score", "", 1, []string{"boxing", "punch", "quick"}, false)
	add("cross", "Cross", "Gyaku-tsuki", "逆突き", GroupStrikes, "strike", "score", "", 1, []string{"boxing", "punch", "reverse"}, false)
	add("hook", "Hook punch", "Hook", "", GroupStrikes, "strike", "score", "", 1, []string{"boxing", "punch"}, false)
	add("uppercut", "Uppercut", "Age-tsuki", "上げ突き", GroupStrikes, "strike", "score", "", 1, []string{"boxing", "punch", "rising"}, false)
	add("overhand", "Overhand punch", "Overhand", "", GroupStrikes, "strike", "score", "", 1, []string{"boxing", "punch", "power"}, false)
	add("backfist", "Backfist strike", "Uraken-uchi", "裏拳打ち", GroupStrikes, "strike", "score", "", 1, []string{"karate", "strike", "reverse"}, false)
	add("straight_punch", "Straight punch", "Choku-tsuki", "直突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "linear"}, false)
	add("lead_straight", "Lead straight punch", "Jun-tsuki", "順突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "lead"}, false)
	add("lunge_punch", "Lunge punch", "Oi-tsuki", "追い突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "stepping"}, false)
	add("snap_punch", "Snap punch", "Kizami-tsuki", "刻み突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "quick"}, false)
	add("body_punch", "Body punch", "Chudan-tsuki", "中段突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "body"}, false)
	add("solar_plexus_punch", "Solar plexus punch", "Suigetsu-tsuki", "水月突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "kyusho"}, false)
	add("spinning_backfist", "Spinning backfist", "Mawashi-uraken", "回し裏拳", GroupStrikes, "strike", "score", "", 1, []string{"karate", "spinning", "strike"}, false)
	add("superman_punch", "Superman punch", "Superman punch", "", GroupStrikes, "strike", "score", "", 1, []string{"kickboxing", "punch", "jumping"}, false)
	add("jumping_punch", "Jumping punch", "Tobi-tsuki", "飛び突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "jumping"}, false)
	add("double_punch", "Double fist punch", "Morote-tsuki", "諸手突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "both hands"}, false)
	add("long_guard_strike", "Long guard straight strike", "Tsukami-tsuki", "掴み突き", GroupStrikes, "strike", "score", "", 1, []string{"jjf", "punch", "guard"}, false)
	add("reversing_punch", "Reversing punch", "Gyaku-zuki", "逆突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "counter"}, false)
	add("swing_punch", "Swinging punch", "Furi-zuki", "振り突き", GroupStrikes, "strike", "score", "", 1, []string{"karate", "punch", "swing"}, false)

	// Kicks -----------------------------------------------------------------
	add("front_kick", "Front kick", "Maegeri", "前蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "linear"}, false)
	add("roundhouse_kick", "Roundhouse kick", "Mawashi-geri", "回し蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "circular"}, false)
	add("side_kick_thrust", "Side kick (thrust)", "Yoko-geri kekomi", "横蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "thrust"}, false)
	add("side_kick_snap", "Side kick (snap)", "Yoko-geri keage", "横蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "snap"}, false)
	add("back_kick", "Back kick", "Ushiro-geri", "後ろ蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "rear"}, false)
	add("spinning_back_kick", "Spinning back kick", "Ushiro-mawashi-geri", "後ろ回し蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "spinning"}, false)
	add("hook_kick", "Hook kick", "Kakato-geri", "踵蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "heel"}, false)
	add("crescent_kick", "Crescent kick", "Mikazuki-geri", "三日月蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "arc"}, false)
	add("wheel_kick", "Wheel kick", "Kuruma-geri", "車蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "spinning"}, false)
	add("axe_kick", "Axe kick", "Ono-geri", "斧蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "downward"}, false)
	add("jumping_front_kick", "Jumping front kick", "Tobi-maegeri", "飛び前蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "jumping"}, false)
	add("jumping_roundhouse", "Jumping roundhouse kick", "Tobi-mawashi-geri", "飛び回し蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "jumping"}, false)
	add("spinning_hook_kick", "Spinning hook kick", "Mawashi-kakato-geri", "回し踵蹴り", GroupKicks, "kick", "score", "", 1, []string{"kickboxing", "kick", "spinning"}, false)
	add("question_mark_kick", "Question mark kick", "Gyaku-mawashi-geri", "逆回し蹴り", GroupKicks, "kick", "score", "", 1, []string{"kickboxing", "kick", "deceptive"}, false)
	add("low_kick", "Low round kick", "Gedan-mawashi-geri", "下段回し蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "low"}, false)
	add("high_kick", "High round kick", "Jodan-mawashi-geri", "上段回し蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "high"}, false)
	add("stomp_kick", "Stomp kick", "Fumi-geri", "踏み蹴り", GroupKicks, "kick", "score", "", 1, []string{"jjf", "kick", "stomp"}, false)
	add("shin_kick", "Shin kick", "Sune-geri", "脛蹴り", GroupKicks, "kick", "score", "", 1, []string{"muay thai", "kick", "shin"}, false)
	add("sweeping_kick", "Sweeping kick", "Nagi-ashi", "薙ぎ足", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "sweep"}, false)
	add("kick_to_body", "Body round kick", "Chudan-mawashi-geri", "中段回し蹴り", GroupKicks, "kick", "score", "", 1, []string{"karate", "kick", "body"}, false)

	// Elbows ----------------------------------------------------------------
	add("horizontal_elbow", "Horizontal elbow strike", "Yoko-hiji-uchi", "横肘打ち", GroupElbows, "strike", "score", "", 1, []string{"muay thai", "elbow", "horizontal"}, false)
	add("upward_elbow", "Upward elbow strike", "Age-hiji-uchi", "上げ肘打ち", GroupElbows, "strike", "score", "", 1, []string{"muay thai", "elbow", "rising"}, false)
	add("downward_elbow", "Downward elbow strike", "Shita-hiji-uchi", "下肘打ち", GroupElbows, "strike", "score", "", 1, []string{"muay thai", "elbow", "downward"}, false)
	add("spinning_elbow", "Spinning elbow strike", "Mawashi-hiji-uchi", "回し肘打ち", GroupElbows, "strike", "score", "", 1, []string{"muay thai", "elbow", "spinning"}, false)
	add("reverse_elbow", "Reverse elbow strike", "Ushiro-hiji-uchi", "後ろ肘打ち", GroupElbows, "strike", "score", "", 1, []string{"muay thai", "elbow", "reverse"}, false)
	add("jumping_elbow", "Jumping elbow strike", "Tobi-hiji-uchi", "飛び肘打ち", GroupElbows, "strike", "score", "", 1, []string{"muay thai", "elbow", "jumping"}, false)
	add("body_elbow", "Elbow strike to body", "Chudan-hiji-uchi", "中段肘打ち", GroupElbows, "strike", "score", "", 1, []string{"muay thai", "elbow", "body"}, false)
	add("head_elbow", "Elbow strike to head", "Jodan-hiji-uchi", "上段肘打ち", GroupElbows, "strike", "score", "", 1, []string{"muay thai", "elbow", "head"}, false)
	add("double_elbow", "Double elbow strike", "Morote-hiji-uchi", "諸手肘打ち", GroupElbows, "strike", "score", "", 1, []string{"jjf", "elbow", "both arms"}, false)
	add("forearm_elbow_smash", "Elbow and forearm smash", "Hiji-haiwan-uchi", "肘腕打ち", GroupElbows, "strike", "score", "", 1, []string{"jjf", "elbow", "forearm"}, false)

	// Knees -----------------------------------------------------------------
	add("straight_knee", "Straight knee strike", "Hiza-uchi", "膝打ち", GroupKnees, "strike", "score", "", 1, []string{"muay thai", "knee", "straight"}, false)
	add("round_knee", "Round knee strike", "Mawashi-hiza-uchi", "回し膝打ち", GroupKnees, "strike", "score", "", 1, []string{"muay thai", "knee", "round"}, false)
	add("jumping_knee", "Jumping knee strike", "Tobi-hiza-uchi", "飛び膝打ち", GroupKnees, "strike", "score", "", 1, []string{"muay thai", "knee", "jumping"}, false)
	add("flying_knee", "Flying knee strike", "Hiza-tobi-uchi", "膝飛び打ち", GroupKnees, "strike", "score", "", 1, []string{"muay thai", "knee", "flying"}, false)
	add("spinning_knee", "Spinning knee strike", "Ushiro-mawashi-hiza", "後ろ回し膝", GroupKnees, "strike", "score", "", 1, []string{"muay thai", "knee", "spinning"}, false)
	add("body_knee", "Knee strike to body", "Chudan-hiza", "中段膝", GroupKnees, "strike", "score", "", 1, []string{"muay thai", "knee", "body"}, false)
	add("head_knee", "Knee strike to head", "Jodan-hiza", "上段膝", GroupKnees, "strike", "score", "", 1, []string{"muay thai", "knee", "head"}, false)
	add("clinch_knee", "Knee from clinch", "Kumi-hiza", "組膝", GroupKnees, "strike", "score", "", 1, []string{"muay thai", "knee", "clinch"}, false)

	// Palm strikes ----------------------------------------------------------
	add("palm_heel_strike", "Palm heel strike", "Teisho-uchi", "掌打", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "palm", "open hand"}, false)
	add("open_hand_slap", "Open hand slap", "Hirate-uchi", "平手打ち", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "palm", "slap"}, false)
	add("spear_hand", "Spear hand strike", "Nukite", "貫手", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "fingers", "kyusho"}, false)
	add("one_finger_spear_hand", "One finger spear hand", "Ippon-nukite", "一本貫手", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "fingers", "kyusho"}, false)
	add("two_finger_spear_hand", "Two finger spear hand", "Nihon-nukite", "二本貫手", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "fingers", "kyusho"}, false)
	add("knife_hand", "Knife hand strike", "Shuto-uchi", "手刀打ち", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "knife hand", "chop"}, false)
	add("ridge_hand", "Ridge hand strike", "Haito-uchi", "背刀打ち", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "ridge hand", "chop"}, false)
	add("forearm_strike", "Forearm strike", "Haiwan-uchi", "腕打ち", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "forearm", "strike"}, false)
	add("cupped_hand_strike", "Cupped hand strike", "Hirate-kuchi", "平手口", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "palm", "cupped"}, false)
	add("ear_clap", "Double ear clap", "Mimi-uchi", "耳打ち", GroupPalmStrikes, "strike", "score", "", 1, []string{"karate", "palm", "double"}, false)

	// Nerve strikes (kyusho) ------------------------------------------------
	add("kyusho_strike", "Kyusho vital point strike", "Kyusho-uchi", "急所打ち", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "vital point", "atemi"}, false)
	add("atemi_strike", "Atemi vital point strike", "Atemi-waza", "当て身", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "vital point", "jjf"}, false)
	add("temple_strike", "Temple strike", "Komekami-uchi", "こめかみ打ち", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "head", "temple"}, false)
	add("throat_strike", "Throat strike", "Nodo-uchi", "喉打ち", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "throat", "nerve"}, false)
	add("solar_plexus_point", "Solar plexus nerve strike", "Suigetsu-uchi", "水月打ち", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "torso", "nerve"}, false)
	add("eye_gouge", "Eye gouge", "Me-tsukami", "目掴み", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "eyes", "dangerous"}, false)
	add("carotid_strike", "Carotid sinus strike", "Keimyaku-uchi", "頸脈打ち", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "neck", "nerve"}, false)
	add("brachial_plexus_strike", "Brachial plexus strike", "Kyusho-ude", "急所腕", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "arm", "nerve"}, false)
	add("radial_nerve_strike", "Radial nerve strike", "Kyusho-kote", "急所小手", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "arm", "nerve"}, false)
	add("peroneal_nerve_strike", "Peroneal nerve strike", "Kyusho-ashi", "急所足", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "leg", "nerve"}, false)
	add("groin_strike", "Groin strike", "Kin-tsuki", "金的突き", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "groin", "dangerous"}, false)
	add("sternum_strike", "Sternum strike", "Mune-uchi", "胸打ち", GroupNerveStrikes, "strike", "score", "", 1, []string{"kyusho", "chest", "nerve"}, false)

	// Throws (te-waza and koshi-waza) ---------------------------------------
	add("seoi_nage", "Shoulder throw", "Seoi-nage", "背負い投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("ippon_seoi_nage", "Single shoulder throw", "Ippon-seoi-nage", "一本背負い投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("morote_seoi_nage", "Two hand shoulder throw", "Morote-seoi-nage", "諸手背負い投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("drop_seoi_nage", "Drop shoulder throw", "Drop seoi-nage", "落とし背負い投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "drop"}, false)
	add("seoi_otoshi", "Shoulder drop throw", "Seoi-otoshi", "背負い落とし", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("kata_guruma", "Shoulder wheel throw", "Kata-guruma", "肩車", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("tai_otoshi", "Body drop throw", "Tai-otoshi", "体落とし", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("sukui_nage", "Scooping throw", "Sukui-nage", "掬い投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("yama_arashi", "Mountain storm throw", "Yama-arashi", "山嵐", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("obi_otoshi", "Belt drop throw", "Obi-otoshi", "帯落とし", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("kata_otoshi", "Shoulder drop", "Kata-otoshi", "肩落とし", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("kaeshi_nage", "Counter throw", "Kaeshi-nage", "返し投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "counter", "nage-waza"}, false)
	add("flying_seoi_nage", "Flying shoulder throw", "Tobi-seoi-nage", "飛び背負い投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "flying"}, false)
	add("gyaku_seoi_nage", "Reverse shoulder throw", "Gyaku-seoi-nage", "逆背負い投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "reverse"}, false)
	add("kata_seoi_nage", "One arm shoulder throw", "Kata-seoi-nage", "片背負い投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "te-waza"}, false)
	add("koshi_guruma", "Hip wheel throw", "Koshi-guruma", "腰車", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("harai_goshi", "Sweeping hip throw", "Harai-goshi", "払腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("tsuri_goshi", "Lifting hip throw", "Tsuri-goshi", "釣腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("tsuri_komi_goshi", "Lifting pulling hip throw", "Tsuri-komi-goshi", "釣込腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("sode_tsuri_komi_goshi", "Sleeve lifting hip throw", "Sode-tsurikomi-goshi", "袖釣込腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("o_goshi", "Major hip throw", "O-goshi", "大腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("uki_goshi", "Floating hip throw", "Uki-goshi", "浮腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("hane_goshi", "Springing hip throw", "Hane-goshi", "跳腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("hanada_goshi", "Hanada hip throw", "Hanada-goshi", "花田腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("uchi_mata", "Inner thigh throw", "Uchi-mata", "内股", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("uchi_goshi", "Inner hip throw", "Uchi-goshi", "内腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("utsuri_goshi", "Changing hip throw", "Utsuri-goshi", "移腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("ushiro_goshi", "Rear hip throw", "Ushiro-goshi", "後腰", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "koshi-waza"}, false)
	add("arm_throw", "Arm throw", "Ude-nage", "腕投げ", GroupThrows, "throw", "score", "", 3, []string{"jjf", "nage-waza", "arm"}, false)
	add("counter_arm_throw", "Counter arm throw", "Ude-gaeshi", "腕返し", GroupThrows, "throw", "score", "", 3, []string{"judo", "counter", "nage-waza"}, false)
	add("dead_body_throw", "Dead body throw", "Shi-sei-nage", "死生投げ", GroupThrows, "throw", "score", "", 3, []string{"judo", "nage-waza", "sacrifice"}, false)
	add("stronger_side_throw", "Throw from stronger grip", "Tsuyo-gata-nage", "強方投げ", GroupThrows, "throw", "score", "", 3, []string{"jjf", "nage-waza", "grip"}, false)
	add("sleeve_turn_throw", "Sleeve turn throw", "Sode-gaeshi", "袖返し", GroupThrows, "throw", "score", "", 3, []string{"judo", "counter", "nage-waza"}, false)
	add("sliding_lapel_throw", "Sliding lapel throw", "Eri-nage", "襟投げ", GroupThrows, "throw", "score", "", 3, []string{"jjf", "nage-waza", "gi"}, false)

	// Takedowns -------------------------------------------------------------
	add("double_leg_takedown", "Double leg takedown", "Morote-gari", "諸手刈", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "bjj", "shot"}, false)
	add("single_leg_takedown", "Single leg takedown", "Single leg", "", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "bjj", "shot"}, false)
	add("high_crotch_takedown", "High crotch takedown", "High crotch", "", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "shot"}, false)
	add("blast_double_takedown", "Blast double takedown", "Blast double", "", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "explosive"}, false)
	add("arm_drag_takedown", "Arm drag takedown", "Ude-hiki taoshi", "腕引き倒し", GroupTakedowns, "takedown", "score", "", 2, []string{"bjj", "arm drag"}, false)
	add("body_lock_takedown", "Body lock takedown", "Daki-taoshi", "抱き倒し", GroupTakedowns, "takedown", "score", "", 2, []string{"jjf", "bjj", "body lock"}, false)
	add("duck_under_takedown", "Duck under takedown", "Duck under", "", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "bjj"}, false)
	add("snap_down", "Snap down", "Snap down", "", GroupTakedowns, "takedown", "score", "", 1, []string{"wrestling", "front headlock"}, false)
	add("front_headlock_takedown", "Front headlock takedown", "Front headlock", "", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "bjj"}, false)
	add("climb_the_ladder", "Climb the ladder takedown", "Climb the ladder", "", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "single leg"}, false)
	add("penetration_shot", "Penetration shot", "Shot", "", GroupTakedowns, "takedown", "score", "", 1, []string{"wrestling", "entry"}, false)
	add("kneeling_shot", "Kneeling shot", "Kneel shot", "", GroupTakedowns, "takedown", "score", "", 1, []string{"wrestling", "entry"}, false)
	add("leg_trip_takedown", "Leg trip takedown", "Ashi-garami", "足緘", GroupTakedowns, "takedown", "score", "", 2, []string{"jjf", "trip"}, false)
	add("whizzer_takedown", "Whizzer takedown", "Whizzer", "", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "overhook"}, false)
	add("russian_tie_takedown", "Russian tie takedown", "Russian tie", "", GroupTakedowns, "takedown", "score", "", 2, []string{"wrestling", "arm drag"}, false)
	add("collar_tie_takedown", "Collar tie takedown", "Collar tie", "", GroupTakedowns, "takedown", "score", "", 1, []string{"wrestling", "clinch"}, false)
	add("overhook_takedown", "Overhook takedown", "Overhook", "", GroupTakedowns, "takedown", "score", "", 1, []string{"wrestling", "bjj"}, false)
	add("underhook_takedown", "Underhook takedown", "Underhook", "", GroupTakedowns, "takedown", "score", "", 1, []string{"wrestling", "bjj"}, false)
	add("crotch_lift_takedown", "Crotch lift takedown", "Kuchiki-taoshi", "朽木倒し", GroupTakedowns, "takedown", "score", "", 2, []string{"judo", "taoshi-waza"}, false)

	// Sweeps ----------------------------------------------------------------
	add("scissor_sweep", "Scissor sweep", "Hasami-sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "guard", "scissors"}, false)
	add("hip_bump_sweep", "Hip bump sweep", "Hip bump", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "guard"}, false)
	add("flower_sweep", "Flower sweep", "Flower sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "guard"}, false)
	add("triangle_sweep", "Triangle sweep", "Sankaku-sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "guard", "triangle"}, false)
	add("butterfly_sweep", "Butterfly sweep", "Butterfly sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "butterfly guard"}, false)
	add("half_guard_sweep", "Half guard sweep", "Half guard sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "half guard"}, false)
	add("pendulum_sweep", "Pendulum sweep", "Pendulum sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "spider guard"}, false)
	add("lasso_sweep", "Lasso sweep", "Lasso sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "lasso guard"}, false)
	add("de_la_riva_sweep", "De la Riva sweep", "De la Riva sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "de la riva"}, false)
	add("spider_guard_sweep", "Spider guard sweep", "Spider sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "spider guard"}, false)
	add("collar_sweep", "Collar and sleeve sweep", "Eri-sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "gi", "collar"}, false)
	add("omoplata_sweep", "Omoplata sweep", "Omoplata sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "omoplata"}, false)
	add("sit_up_sweep", "Sit-up sweep", "Sit-up sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "guard"}, false)
	add("underhook_sweep", "Underhook sweep", "Underhook sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "half guard"}, false)
	add("arm_in_guard_sweep", "Arm-in guard sweep", "Arm-in sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "guard"}, false)
	add("reverse_de_la_riva_sweep", "Reverse De la Riva sweep", "Reverse De la Riva", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "de la riva"}, false)
	add("rubber_guard_sweep", "Rubber guard sweep", "Rubber guard sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "rubber guard"}, false)
	add("worm_guard_sweep", "Worm guard sweep", "Worm guard sweep", "", GroupSweeps, "sweep", "score", "", 2, []string{"bjj", "worm guard"}, false)

	// Reaps -----------------------------------------------------------------
	add("ouchi_gari", "Inside reap", "Ouchi-gari", "大内刈", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "reap"}, false)
	add("kouchi_gari", "Minor inside reap", "Kouchi-gari", "小内刈", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "reap"}, false)
	add("osoto_gari", "Major outer reap", "Osoto-gari", "大外刈", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "reap"}, false)
	add("kosoto_gari", "Minor outer reap", "Kosoto-gari", "小外刈", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "reap"}, false)
	add("ashi_guruma", "Foot wheel throw", "Ashi-guruma", "足車", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "wheel"}, false)
	add("osoto_otoshi", "Major outer drop", "Osoto-otoshi", "大外落とし", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "drop"}, false)
	add("osoto_makikomi", "Major outer winding throw", "Osoto-makikomi", "大外巻込", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "makikomi"}, false)
	add("sasae_tsuri_komi_ashi", "Blocking foot pull throw", "Sasae-tsurikomi-ashi", "嵐脚釣込足", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "block"}, false)
	add("hiza_guruma", "Knee wheel throw", "Hiza-guruma", "膝車", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "wheel"}, false)
	add("o_guruma", "Great wheel throw", "O-guruma", "大車", GroupReaps, "reap", "score", "", 2, []string{"judo", "ashi-waza", "wheel"}, false)

	// Hooks -----------------------------------------------------------------
	add("kosoto_gake", "Minor outer hook", "Kosoto-gake", "小外掛", GroupHooks, "hook", "score", "", 2, []string{"judo", "ashi-waza", "hook"}, false)
	add("kouchi_gake", "Minor inner hook", "Kouchi-gake", "小内掛", GroupHooks, "hook", "score", "", 2, []string{"judo", "ashi-waza", "hook"}, false)
	add("leg_hook_trip", "Leg hook trip (Ashi-garami)", "Ashi-garami", "足緘", GroupHooks, "hook", "score", "", 2, []string{"jjf", "hook", "trip"}, false)
	add("leg_hook", "Leg hook", "Ashi-kake", "足掛", GroupHooks, "hook", "score", "", 1, []string{"jjf", "hook", "leg"}, false)
	add("overhook_trip", "Overhook trip", "Whizzer trip", "", GroupHooks, "hook", "score", "", 2, []string{"wrestling", "overhook", "trip"}, false)
	add("underhook_trip", "Underhook trip", "Underhook trip", "", GroupHooks, "hook", "score", "", 2, []string{"wrestling", "underhook", "trip"}, false)
	add("hook_throw", "Hook throw", "Kagi-nage", "鉤投げ", GroupHooks, "hook", "score", "", 2, []string{"jjf", "hook", "throw"}, false)
	add("inside_hook_trip", "Inside hook trip", "Uchi-gake", "内掛", GroupHooks, "hook", "score", "", 2, []string{"jjf", "hook", "trip"}, false)

	// Sacrifice techniques (sutemi-waza) ------------------------------------
	add("tomoe_nage", "Circle throw", "Tomoe-nage", "巴投げ", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "sacrifice"}, false)
	add("ura_nage", "Rear throw", "Ura-nage", "裏投げ", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "sacrifice"}, false)
	add("yoko_guruma", "Side wheel throw", "Yoko-guruma", "横車", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "sacrifice"}, false)
	add("yoko_tomoe_nage", "Side circle throw", "Yoko-tomoe-nage", "横巴投げ", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "sacrifice"}, false)
	add("tani_otoshi", "Valley drop", "Tani-otoshi", "谷落とし", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "sacrifice"}, false)
	add("sumi_otoshi", "Corner drop", "Sumi-otoshi", "隅落とし", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "sacrifice"}, false)
	add("hikikomi_gaeshi", "Pulling sacrifice counter", "Hikikomi-gaeshi", "引き込み返し", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "counter"}, false)
	add("tawara_gaeshi", "Rice bag counter", "Tawara-gaeshi", "俵返", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "counter"}, false)
	add("yoko_gake", "Side sacrifice throw", "Yoko-gake", "横掛", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "sacrifice"}, false)
	add("uchi_makikomi", "Inner winding throw", "Uchi-makikomi", "内巻込", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "makikomi"}, false)
	add("soto_makikomi", "Outer winding throw", "Soto-makikomi", "外巻込", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "makikomi"}, false)
	add("hane_makikomi", "Springing winding throw", "Hane-makikomi", "跳巻込", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sutemi-waza", "makikomi"}, false)
	add("kani_basami", "Scissors throw", "Kani-basami", "蟹挟", GroupSacrifice, "throw", "score", "", 0, []string{"judo", "prohibited", "sacrifice"}, false)
	add("sutemi_waza", "Sacrifice technique", "Sutemi-waza", "捨て身技", GroupSacrifice, "throw", "score", "", 3, []string{"judo", "sacrifice", "generic"}, false)
	add("flying_tackle_throw", "Flying tackle throw", "Tobi-taoshi", "飛び倒し", GroupSacrifice, "throw", "score", "", 3, []string{"jjf", "sacrifice", "flying"}, false)
	add("scissors_takedown", "Scissors takedown", "Hasami-taoshi", "挟み倒し", GroupSacrifice, "throw", "score", "", 2, []string{"jjf", "sacrifice", "scissors"}, false)

	// Ground fighting -------------------------------------------------------
	add("guard_pull", "Guard pull", "Guard pull", "", GroupGroundFighting, "ground", "score", "", 0, []string{"bjj", "guard", "entry"}, false)
	add("sit_down_guard", "Sit down to guard", "Zagi-gaeshi", "座り返し", GroupGroundFighting, "ground", "score", "", 0, []string{"bjj", "guard", "entry"}, false)
	add("sprawl", "Sprawl", "Sprawl", "", GroupGroundFighting, "ground", "score", "", 1, []string{"wrestling", "bjj", "defence"}, false)
	add("turtle_position", "Turtle position", "Turtle", "", GroupGroundFighting, "ground", "score", "", 0, []string{"bjj", "defence", "turtle"}, false)
	add("bear_crawl", "Bear crawl", "Bear crawl", "", GroupGroundFighting, "ground", "score", "", 0, []string{"bjj", "movement"}, false)
	add("scramble", "Scramble", "Scramble", "", GroupGroundFighting, "ground", "score", "", 0, []string{"bjj", "wrestling", "movement"}, false)
	add("granby_roll", "Granby roll", "Granby roll", "", GroupGroundFighting, "ground", "score", "", 0, []string{"wrestling", "bjj", "roll"}, false)
	add("forward_roll", "Forward roll", "Mae-kaiten", "前回転", GroupGroundFighting, "ground", "score", "", 0, []string{"judo", "ukemi", "roll"}, false)
	add("backward_roll", "Backward roll", "Ushiro-kaiten", "後回転", GroupGroundFighting, "ground", "score", "", 0, []string{"judo", "ukemi", "roll"}, false)
	add("seated_guard_entry", "Seated guard entry", "Seated guard", "", GroupGroundFighting, "ground", "score", "", 0, []string{"bjj", "guard", "entry"}, false)
	add("deep_half_entry", "Deep half guard entry", "Deep half entry", "", GroupGroundFighting, "ground", "score", "", 0, []string{"bjj", "half guard", "entry"}, false)
	add("sliding_guard_entry", "Sliding guard entry", "Sliding guard", "", GroupGroundFighting, "ground", "score", "", 0, []string{"bjj", "guard", "entry"}, false)

	// Hold downs ------------------------------------------------------------
	add("osae_komi", "Hold down", "Osae-komi", "押さえ込み", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("kesa_gatame", "Scarf hold", "Kesa-gatame", "袈裟固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("kuzure_kesa_gatame", "Broken scarf hold", "Kuzure-kesa-gatame", "崩袈裟固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("makura_kesa_gatame", "Pillow scarf hold", "Makura-kesa-gatame", "枕袈裟固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("yoko_shiho_gatame", "Side four corner hold", "Yoko-shiho-gatame", "横四方固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("kuzure_yoko_shiho_gatame", "Broken side four corner hold", "Kuzure-yoko-shiho-gatame", "崩横四方固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("tate_shiho_gatame", "Vertical four corner hold", "Tate-shiho-gatame", "縦四方固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("kuzure_tate_shiho_gatame", "Broken vertical four corner hold", "Kuzure-tate-shiho-gatame", "崩縦四方固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("kami_shiho_gatame", "Upper four corner hold", "Kami-shiho-gatame", "上四方固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("kuzure_kami_shiho_gatame", "Broken upper four corner hold", "Kuzure-kami-shiho-gatame", "崩上四方固", GroupHoldDowns, "hold", "score", "", 1, []string{"judo", "ne-waza", "hold"}, false)
	add("osae_komi_wazaari", "Hold down ten seconds (Waza-ari)", "Osae-komi 10s", "押さえ込み十秒", GroupHoldDowns, "wazaari", "score", "", 1, []string{"judo", "hold timer", "waza-ari"}, false)
	add("osae_komi_ippon", "Hold down twenty seconds (Ippon)", "Osae-komi 20s", "押さえ込み二十秒", GroupHoldDowns, "ippon", "score", "", 3, []string{"judo", "hold timer", "ippon"}, false)

	// Pins ------------------------------------------------------------------
	add("side_control_pin", "Side control pin", "Side control", "", GroupPins, "hold", "score", "", 1, []string{"bjj", "pin", "side control"}, false)
	add("mount_pin", "Mount pin", "Mount", "", GroupPins, "hold", "score", "", 1, []string{"bjj", "pin", "mount"}, false)
	add("back_pin", "Back control pin", "Back control", "", GroupPins, "hold", "score", "", 1, []string{"bjj", "pin", "back"}, false)
	add("knee_on_belly_pin", "Knee on belly pin", "Knee-on-belly", "", GroupPins, "hold", "score", "", 2, []string{"bjj", "pin", "knee on belly"}, false)
	add("north_south_pin", "North-south pin", "North-south", "", GroupPins, "hold", "score", "", 1, []string{"bjj", "pin"}, false)
	add("reverse_side_control_pin", "Reverse side control pin", "Reverse side control", "", GroupPins, "hold", "score", "", 1, []string{"bjj", "pin", "side control"}, false)
	add("scarf_hold_pin", "Scarf hold pin", "Scarf hold", "", GroupPins, "hold", "score", "", 1, []string{"bjj", "pin", "kesa"}, false)
	add("head_and_arm_pin", "Head and arm pin", "Head and arm", "", GroupPins, "hold", "score", "", 1, []string{"wrestling", "pin"}, false)
	add("body_lock_pin", "Body lock pin", "Body lock", "", GroupPins, "hold", "score", "", 1, []string{"bjj", "pin", "body lock"}, false)
	add("turtle_pin", "Turtle pin", "Turtle pin", "", GroupPins, "hold", "score", "", 1, []string{"bjj", "pin", "turtle"}, false)

	// Submissions -----------------------------------------------------------
	add("tap_out", "Tap out", "Tap-out", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "result", "tap"}, true)
	add("verbal_tap", "Verbal tap", "Verbal tap", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "result", "tap"}, true)
	add("armbar", "Armbar", "Juji-gatame", "十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "judo", "arm lock", "instant win"}, true)
	add("straight_armbar", "Straight armbar", "Juji-gatame (straight)", "十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "arm lock", "instant win"}, true)
	add("bent_armbar", "Bent armbar", "Juji-gatame (bent)", "十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "arm lock", "instant win"}, true)
	add("flying_armbar", "Flying armbar", "Tobi-juji-gatame", "飛び十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "arm lock", "flying"}, true)
	add("spinning_armbar", "Spinning armbar", "Mawashi-juji-gatame", "回し十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "arm lock", "spinning"}, true)
	add("inverted_armbar", "Inverted armbar", "Saka-juji-gatame", "逆十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "arm lock", "inverted"}, true)
	add("guard_armbar", "Armbar from guard", "Guard juji-gatame", "ガード十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "arm lock", "guard"}, true)
	add("kimura", "Kimura lock", "Ude-garami", "腕緘", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "judo", "shoulder lock", "instant win"}, true)
	add("reverse_kimura", "Reverse kimura", "Gyaku-ude-garami", "逆腕緘", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "shoulder lock", "instant win"}, true)
	add("americana", "Americana", "Americana", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "shoulder lock", "instant win"}, true)
	add("keylock", "Keylock", "Keylock", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "shoulder lock", "instant win"}, true)
	add("omoplata", "Omoplata", "Omoplata", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "shoulder lock", "instant win"}, true)
	add("omoplata_roll", "Omoplata roll", "Omoplata roll", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "shoulder lock", "roll"}, true)
	add("omoplata_to_triangle", "Omoplata to triangle", "Omoplata sankaku", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "combination", "triangle"}, true)
	add("shoulder_lock", "Shoulder lock", "Kata-gatame", "肩固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "judo", "shoulder lock"}, true)
	add("heel_hook", "Heel hook", "Heel hook", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "instant win"}, true)
	add("inside_heel_hook", "Inside heel hook", "Inside heel hook", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "heel hook"}, true)
	add("outside_heel_hook", "Outside heel hook", "Outside heel hook", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "heel hook"}, true)
	add("toe_hold_heel_hook", "Toe hold heel hook", "Toe hold hook", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "heel hook"}, true)
	add("kneebar", "Kneebar", "Hiza-gatame", "膝固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "judo", "leg lock", "instant win"}, true)
	add("reverse_kneebar", "Reverse kneebar", "Gyaku-hiza-gatame", "逆膝固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "kneebar"}, true)
	add("ankle_lock", "Ankle lock", "Ashi-gatame", "足固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "instant win"}, true)
	add("toe_hold", "Toe hold", "Toe hold", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "instant win"}, true)
	add("calf_slicer", "Calf slicer", "Calf slicer", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "instant win"}, true)
	add("leg_bar", "Leg bar", "Ashi-juji-gatame", "足十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "instant win"}, true)
	add("figure_four_leglock", "Figure four leg lock", "Figure-four", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "instant win"}, true)
	add("banana_split", "Banana split", "Banana split", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "instant win"}, true)
	add("wrist_lock", "Wrist lock", "Kote-garami", "小手緘", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "wrist lock", "instant win"}, true)
	add("finger_lock", "Finger lock", "Yubi-garami", "指緘", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "finger lock", "instant win"}, true)
	add("triangle_armbar", "Triangle armbar", "Sankaku-juji-gatame", "三角十字固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "combination", "triangle"}, true)
	add("triangle_omoplata", "Triangle omoplata", "Sankaku-omoplata", "", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "combination", "triangle"}, true)
	add("guard_wrist_lock", "Wrist lock from guard", "Guard kote-garami", "ガード小手緘", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "wrist lock", "guard"}, true)
	add("mount_shoulder_lock", "Shoulder lock from mount", "Mount kata-gatame", "マウント肩固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "shoulder lock", "mount"}, true)
	add("flying_kneebar", "Flying kneebar", "Tobi-hiza-gatame", "飛び膝固", GroupSubmissions, "submission", "submission", "", 0, []string{"bjj", "leg lock", "flying"}, true)
	add("ippon", "Ippon (full point)", "Ippon", "一本", GroupSubmissions, "ippon", "score", "", 3, []string{"result", "score", "instant win"}, false)
	add("wazaari", "Waza-ari (half point)", "Waza-ari", "技あり", GroupSubmissions, "wazaari", "score", "", 1, []string{"result", "score"}, false)

	// Chokes ----------------------------------------------------------------
	add("rear_naked_choke", "Rear naked choke", "Hadaka-jime", "裸絞", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "judo", "choke", "instant win"}, true)
	add("body_lock_choke", "Body lock choke", "Hadaka-jime (body lock)", "裸絞", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "body lock"}, true)
	add("triangle_choke", "Triangle choke", "Sankaku-jime", "三角絞", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "judo", "choke", "instant win"}, true)
	add("arm_in_triangle", "Arm-in triangle choke", "Sankaku-jime (arm-in)", "三角絞", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "triangle"}, true)
	add("guillotine_choke", "Guillotine choke", "Guillotine", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "front headlock", "instant win"}, true)
	add("arm_in_guillotine", "Arm-in guillotine", "Arm-in guillotine", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "guillotine"}, true)
	add("darce_choke", "D'Arce choke", "D'Arce", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "front headlock", "instant win"}, true)
	add("brabo_choke", "Brabo choke", "Brabo", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "darce"}, true)
	add("anaconda_choke", "Anaconda choke", "Anaconda", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "front headlock", "instant win"}, true)
	add("gogoplata", "Gogoplata", "Gogoplata", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "instant win"}, true)
	add("clock_choke", "Clock choke", "Koshi-jime", "腰絞", GroupChokes, "submission", "submission", "", 0, []string{"judo", "choke", "instant win"}, true)
	add("bow_and_arrow_choke", "Bow and arrow choke", "Bow and arrow", "", GroupChokes, "submission", "submission", "", 0, []string{"judo", "bjj", "choke", "gi"}, true)
	add("okuri_eri_jime", "Sliding collar choke", "Okuri-eri-jime", "送襟絞", GroupChokes, "submission", "submission", "", 0, []string{"judo", "choke", "gi"}, true)
	add("kata_jime", "Single collar choke", "Kata-jime", "片絞", GroupChokes, "submission", "submission", "", 0, []string{"judo", "choke", "gi"}, true)
	add("nami_juji_jime", "Normal cross choke", "Nami-juji-jime", "並十字絞", GroupChokes, "submission", "submission", "", 0, []string{"judo", "choke", "gi"}, true)
	add("gyaku_juji_jime", "Reverse cross choke", "Gyaku-juji-jime", "逆十字絞", GroupChokes, "submission", "submission", "", 0, []string{"judo", "choke", "gi"}, true)
	add("kataha_jime", "Half collar choke", "Kataha-jime", "片羽絞", GroupChokes, "submission", "submission", "", 0, []string{"judo", "choke", "gi"}, true)
	add("ryote_jime", "Both hand choke", "Ryote-jime", "両手絞", GroupChokes, "submission", "submission", "", 0, []string{"judo", "choke"}, true)
	add("tsukkomi_jime", "Thrust choke", "Tsukkomi-jime", "突込絞", GroupChokes, "submission", "submission", "", 0, []string{"judo", "choke", "gi"}, true)
	add("front_headlock_choke", "Front headlock choke", "Front headlock", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "front headlock"}, true)
	add("peruvian_necktie", "Peruvian necktie", "Peruvian necktie", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "guillotine"}, true)
	add("chin_strap_choke", "Chin strap choke", "Chin strap", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "front headlock"}, true)
	add("ezekiel_choke", "Ezekiel choke", "Ezekiel", "", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "choke", "gi", "sleeve"}, true)
	add("sleeve_choke", "Sleeve choke", "Sode-jime", "袖絞", GroupChokes, "submission", "submission", "", 0, []string{"bjj", "judo", "choke", "gi"}, true)

	// Joint locks (kansetsu-waza and modern variants) -----------------------
	add("waki_gatame", "Armpit lock", "Waki-gatame", "脇固", GroupJointLocks, "submission", "submission", "", 0, []string{"judo", "joint lock", "arm"}, true)
	add("kata_gatame_lock", "Shoulder lock (Kata-gatame)", "Kata-gatame", "肩固", GroupJointLocks, "submission", "submission", "", 0, []string{"judo", "joint lock", "shoulder"}, true)
	add("ashi_gatame_lock", "Leg lock (Ashi-gatame)", "Ashi-gatame", "足固", GroupJointLocks, "submission", "submission", "", 0, []string{"judo", "joint lock", "leg"}, true)
	add("do_jime", "Body lock (Do-jime)", "Do-jime", "胴絞", GroupJointLocks, "submission", "submission", "", 0, []string{"judo", "prohibited", "body lock"}, false)
	add("elbow_lock", "Elbow lock", "Hiji-gatame", "肘固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "joint lock", "elbow"}, true)
	add("wrist_crank", "Wrist crank", "Kote-gatame", "小手固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "joint lock", "wrist"}, true)
	add("finger_crank", "Finger crank", "Yubi-gatame", "指固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "joint lock", "finger"}, true)
	add("toe_crank", "Toe crank", "Tsuma-gatame", "爪固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "joint lock", "toe"}, true)
	add("ankle_crank", "Ankle crank", "Ashi-kubi-gatame", "足首固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "joint lock", "ankle"}, true)
	add("shoulder_crank", "Shoulder crank", "Kata-gatame (crank)", "肩固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "joint lock", "shoulder"}, true)
	add("neck_crank", "Neck crank", "Kubi-gatame", "首固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "prohibited", "neck"}, false)
	add("spine_lock", "Spine lock", "Spine lock", "", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "prohibited", "spine"}, false)
	add("figure_four_armlock", "Figure four arm lock", "Figure-four arm lock", "", GroupJointLocks, "submission", "submission", "", 0, []string{"bjj", "joint lock", "arm"}, true)
	add("straight_ankle_lock", "Straight ankle lock", "Straight ankle lock", "", GroupJointLocks, "submission", "submission", "", 0, []string{"bjj", "joint lock", "ankle"}, true)
	add("estilo_lock", "Estilo lock", "Estilo lock", "", GroupJointLocks, "submission", "submission", "", 0, []string{"bjj", "joint lock", "leg"}, true)
	add("calf_crank", "Calf crank", "Calf crank", "", GroupJointLocks, "submission", "submission", "", 0, []string{"bjj", "joint lock", "leg"}, true)
	add("knee_lock", "Knee lock", "Hiza-gatame", "膝固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "joint lock", "knee"}, true)
	add("hip_lock", "Hip lock", "Koshi-gatame", "腰固", GroupJointLocks, "submission", "submission", "", 0, []string{"jjf", "joint lock", "hip"}, true)

	// Controls --------------------------------------------------------------
	add("closed_guard", "Closed guard", "Closed guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "control"}, false)
	add("open_guard", "Open guard", "Open guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "control"}, false)
	add("half_guard", "Half guard", "Half guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "control"}, false)
	add("butterfly_guard", "Butterfly guard", "Butterfly guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "control"}, false)
	add("spider_guard", "Spider guard", "Spider guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "gi"}, false)
	add("de_la_riva_guard", "De la Riva guard", "De la Riva", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "gi"}, false)
	add("reverse_de_la_riva_guard", "Reverse De la Riva guard", "Reverse De la Riva", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "gi"}, false)
	add("lasso_guard", "Lasso guard", "Lasso guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "gi"}, false)
	add("worm_guard", "Worm guard", "Worm guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "gi"}, false)
	add("rubber_guard", "Rubber guard", "Rubber guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "no-gi"}, false)
	add("floating_guard", "Floating guard", "Floating guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "control"}, false)
	add("seated_guard", "Seated guard", "Seated guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "control"}, false)
	add("sit_up_guard", "Sit-up guard", "Sit-up guard", "", GroupControls, "control", "score", "", 0, []string{"bjj", "guard", "control"}, false)
	add("mount_control", "Mount", "Mount", "", GroupControls, "control", "score", "", 2, []string{"bjj", "control", "mount"}, false)
	add("back_control", "Back control", "Back control", "", GroupControls, "control", "score", "", 2, []string{"bjj", "control", "back"}, false)
	add("side_control", "Side control", "Side control", "", GroupControls, "control", "score", "", 1, []string{"bjj", "control", "side control"}, false)
	add("north_south_control", "North-south position", "North-south", "", GroupControls, "control", "score", "", 0, []string{"bjj", "control"}, false)
	add("knee_on_belly_control", "Knee on belly control", "Knee-on-belly", "", GroupControls, "control", "score", "", 2, []string{"bjj", "control", "knee on belly"}, false)
	add("turtle_guard", "Turtle guard", "Turtle", "", GroupControls, "control", "score", "", 0, []string{"bjj", "defence", "turtle"}, false)
	add("front_headlock_control", "Front headlock", "Front headlock", "", GroupControls, "control", "score", "", 0, []string{"bjj", "wrestling", "control"}, false)
	add("body_lock_control", "Body lock", "Body lock", "", GroupControls, "control", "score", "", 0, []string{"bjj", "back control", "clinch"}, false)
	add("overhook_control", "Overhook / whizzer", "Overhook", "", GroupControls, "control", "score", "", 0, []string{"wrestling", "bjj", "clinch"}, false)
	add("underhook_control", "Underhook", "Underhook", "", GroupControls, "control", "score", "", 0, []string{"wrestling", "bjj", "clinch"}, false)
	add("collar_tie_control", "Collar tie", "Collar tie", "", GroupControls, "control", "score", "", 0, []string{"wrestling", "clinch", "gi"}, false)

	// Breaks and escapes ----------------------------------------------------
	add("bridge_escape", "Bridge", "Bridge", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "bridge"}, false)
	add("bridge_and_roll", "Bridge and roll escape", "Bridge and roll", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "mount"}, false)
	add("shrimp_escape", "Shrimp / hip escape", "Shrimp", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "hip"}, false)
	add("sit_out_escape", "Sit-out", "Sit-out", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "guard"}, false)
	add("technical_stand_up", "Technical stand-up", "Technical stand-up", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "get up"}, false)
	add("guard_break", "Guard break", "Guard break", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "guard"}, false)
	add("posture_up", "Posture up", "Posture", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "guard"}, false)
	add("frame_escape", "Frame escape", "Frame", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "frame"}, false)
	add("pummel", "Pummel / hand fighting", "Pummel", "", GroupBreaks, "break", "score", "", 0, []string{"wrestling", "bjj", "clinch"}, false)
	add("armbar_escape", "Armbar escape", "Juji-gatame escape", "十字固逃げ", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "arm lock"}, false)
	add("triangle_escape", "Triangle escape", "Sankaku-jime escape", "三角絞逃げ", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "triangle"}, false)
	add("rear_naked_choke_escape", "Rear naked choke escape", "Hadaka-jime escape", "裸絞逃げ", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "choke"}, false)
	add("guillotine_escape", "Guillotine escape", "Guillotine escape", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "choke"}, false)
	add("mount_escape", "Mount escape", "Mount escape", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "mount"}, false)
	add("side_control_escape", "Side control escape", "Side control escape", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "side control"}, false)
	add("back_control_escape", "Back control escape", "Back escape", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "back"}, false)
	add("half_guard_escape", "Half guard escape", "Half guard escape", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "half guard"}, false)
	add("turtle_escape", "Turtle escape", "Turtle escape", "", GroupBreaks, "break", "score", "", 0, []string{"bjj", "escape", "turtle"}, false)

	// Standing variants -----------------------------------------------------
	add("standing_armbar", "Standing armbar", "Tachi-juji-gatame", "立ち十字固", GroupStandingVariants, "submission", "submission", "", 0, []string{"jjf", "standing", "arm lock"}, true)
	add("standing_guillotine", "Standing guillotine", "Tachi-guillotine", "立ちギロチン", GroupStandingVariants, "submission", "submission", "", 0, []string{"jjf", "standing", "choke"}, true)
	add("standing_rear_naked_choke", "Standing rear naked choke", "Tachi-hadaka-jime", "立ち裸絞", GroupStandingVariants, "submission", "submission", "", 0, []string{"jjf", "standing", "choke"}, true)
	add("standing_kimura", "Standing kimura", "Tachi-ude-garami", "立ち腕緘", GroupStandingVariants, "submission", "submission", "", 0, []string{"jjf", "standing", "shoulder lock"}, true)
	add("standing_headlock", "Standing front headlock", "Tachi-headlock", "立ちヘッドロック", GroupStandingVariants, "control", "score", "", 0, []string{"wrestling", "standing", "headlock"}, false)
	add("standing_ankle_lock", "Standing ankle lock", "Tachi-ashi-gatame", "立ち足固", GroupStandingVariants, "submission", "submission", "", 0, []string{"jjf", "standing", "leg lock"}, true)
	add("flying_guillotine", "Flying guillotine", "Tobi-guillotine", "飛びギロチン", GroupStandingVariants, "submission", "submission", "", 0, []string{"jjf", "flying", "choke"}, true)
	add("jumping_choke", "Jumping choke", "Tobi-jime", "飛び絞", GroupStandingVariants, "submission", "submission", "", 0, []string{"jjf", "jumping", "choke"}, true)
	add("standing_wrist_lock", "Standing wrist lock", "Tachi-kote-garami", "立ち小手緘", GroupStandingVariants, "submission", "submission", "", 0, []string{"jjf", "standing", "wrist lock"}, true)
	add("standing_counter_throw", "Standing counter throw", "Tachi-kaeshi-nage", "立ち返し投げ", GroupStandingVariants, "throw", "score", "", 3, []string{"jjf", "standing", "counter"}, false)

	// Ground variants -------------------------------------------------------
	add("ground_punch", "Ground punch", "Ne-waza tsuki", "寝技突き", GroupGroundVariants, "strike", "score", "", 1, []string{"jjf", "ground", "punch"}, false)
	add("ground_elbow", "Ground elbow strike", "Ne-waza hiji-uchi", "寝技肘打ち", GroupGroundVariants, "strike", "score", "", 1, []string{"jjf", "ground", "elbow"}, false)
	add("ground_knee", "Ground knee strike", "Ne-waza hiza-uchi", "寝技膝打ち", GroupGroundVariants, "strike", "score", "", 1, []string{"jjf", "ground", "knee"}, false)
	add("ground_kick", "Ground kick", "Ne-waza geri", "寝技蹴り", GroupGroundVariants, "strike", "score", "", 1, []string{"jjf", "ground", "kick"}, false)
	add("ground_headbutt", "Ground headbutt", "Ne-waza atama-uchi", "寝技頭打ち", GroupGroundVariants, "strike", "score", "", 1, []string{"jjf", "ground", "headbutt"}, false)
	add("ground_palm_strike", "Ground palm strike", "Ne-waza teisho-uchi", "寝技掌打", GroupGroundVariants, "strike", "score", "", 1, []string{"jjf", "ground", "palm"}, false)
	add("ground_knife_hand", "Ground knife hand strike", "Ne-waza shuto-uchi", "寝技手刀打ち", GroupGroundVariants, "strike", "score", "", 1, []string{"jjf", "ground", "knife hand"}, false)
	add("ground_eye_gouge", "Ground eye gouge", "Ne-waza kyusho-me", "寝技急所目", GroupGroundVariants, "strike", "score", "", 0, []string{"jjf", "ground", "kyusho", "dangerous"}, false)
	add("ground_throat_strike", "Ground throat strike", "Ne-waza nodo-uchi", "寝技喉打ち", GroupGroundVariants, "strike", "score", "", 0, []string{"jjf", "ground", "kyusho", "dangerous"}, false)
	add("ground_forearm_smash", "Ground forearm smash", "Ne-waza haiwan-uchi", "寝技腕打ち", GroupGroundVariants, "strike", "score", "", 1, []string{"jjf", "ground", "forearm"}, false)

	// Kata ------------------------------------------------------------------
	add("heian_shodan", "Heian Shodan", "Heian shodan", "平安初段", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "shitei"}, false)
	add("heian_nidan", "Heian Nidan", "Heian nidan", "平安二段", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "shitei"}, false)
	add("heian_sandan", "Heian Sandan", "Heian sandan", "平安三段", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "shitei"}, false)
	add("heian_yondan", "Heian Yondan", "Heian yondan", "平安四段", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "shitei"}, false)
	add("heian_godan", "Heian Godan", "Heian godan", "平安五段", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "shitei"}, false)
	add("tekki_shodan", "Tekki Shodan", "Tekki shodan", "鉄騎初段", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "shitei"}, false)
	add("tekki_nidan", "Tekki Nidan", "Tekki nidan", "鉄騎二段", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "tokuyi"}, false)
	add("tekki_sandan", "Tekki Sandan", "Tekki sandan", "鉄騎三段", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "tokuyi"}, false)
	add("bassai_dai", "Bassai Dai", "Bassai dai", "抜塞大", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "shitei"}, false)
	add("bassai_sho", "Bassai Sho", "Bassai sho", "抜塞小", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "tokuyi"}, false)
	add("kanku_dai", "Kanku Dai", "Kanku dai", "観空大", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "shitei"}, false)
	add("kanku_sho", "Kanku Sho", "Kanku sho", "観空小", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "tokuyi"}, false)
	add("nage_no_kata", "Nage-no-kata (forms of throwing)", "Nage-no-kata", "投の形", GroupKata, "kata", "score", "", 0, []string{"judo", "kata", "throws"}, false)
	add("katame_no_kata", "Katame-no-kata (forms of holding)", "Katame-no-kata", "固の形", GroupKata, "kata", "score", "", 0, []string{"judo", "kata", "holds"}, false)
	add("kime_no_kata", "Kime-no-kata (forms of fighting)", "Kime-no-kata", "極の形", GroupKata, "kata", "score", "", 0, []string{"judo", "kata", "self defence"}, false)
	add("ju_no_kata", "Ju-no-kata (forms of gentleness)", "Ju-no-kata", "柔の形", GroupKata, "kata", "score", "", 0, []string{"judo", "kata", "principles"}, false)
	add("goshin_jutsu", "Kodokan Goshin-jutsu", "Goshin-jutsu", "護身術", GroupKata, "kata", "score", "", 0, []string{"judo", "kata", "self defence"}, false)
	add("bunkai", "Bunkai application", "Bunkai", "分解", GroupKata, "kata", "score", "", 0, []string{"karate", "kata", "application"}, false)
	add("bo_kata", "Bo staff kata", "Bo-kata", "棒型", GroupKata, "kata", "score", "", 0, []string{"weapon kata", "bo", "staff"}, false)
	add("sai_kata", "Sai kata", "Sai-kata", "釵型", GroupKata, "kata", "score", "", 0, []string{"weapon kata", "sai"}, false)
	add("nunchaku_kata", "Nunchaku kata", "Nunchaku-kata", "ヌンチャク型", GroupKata, "kata", "score", "", 0, []string{"weapon kata", "nunchaku"}, false)
	add("kama_kata", "Kama kata", "Kama-kata", "鎌型", GroupKata, "kata", "score", "", 0, []string{"weapon kata", "kama", "sickle"}, false)

	// Penalties -------------------------------------------------------------
	add("yellow_card", "Yellow card / foul", "Shido", "", GroupPenalties, "penalty", "penalty", "yellow", 0, []string{"card", "foul", "shido"}, false)
	add("red_card", "Red card / serious foul", "Hansoku", "", GroupPenalties, "penalty", "penalty", "red", 0, []string{"card", "foul", "hansoku"}, false)
	add("shido_penalty", "Shido (light penalty)", "Shido", "指導", GroupPenalties, "penalty", "penalty", "shido", 0, []string{"card", "judo", "foul"}, false)
	add("two_shidos", "Two shidos", "Ni-shido", "二指導", GroupPenalties, "penalty", "penalty", "shido", 0, []string{"card", "judo", "foul"}, false)
	add("hansoku_make", "Hansoku-make (disqualification)", "Hansoku-make", "反則負け", GroupPenalties, "penalty", "penalty", "hansoku", 0, []string{"card", "judo", "disqualification"}, false)
	add("direct_hansoku_make", "Direct hansoku-make", "Direct hansoku-make", "直接反則負け", GroupPenalties, "penalty", "penalty", "hansoku", 0, []string{"card", "judo", "disqualification"}, false)
	add("forfeit", "Forfeit", "Forfeit", "", GroupPenalties, "penalty", "penalty", "", 0, []string{"withdrawal", "result", "no contest"}, false)
	add("injury_withdrawal", "Injury withdrawal", "Injury", "", GroupPenalties, "penalty", "penalty", "", 0, []string{"injury", "withdrawal", "result"}, false)
	add("no_show", "No-show", "No-show", "", GroupPenalties, "penalty", "penalty", "", 0, []string{"withdrawal", "attendance", "result"}, false)
	add("cheating", "Cheating / disqualification", "Cheating", "", GroupPenalties, "penalty", "penalty", "red", 0, []string{"disqualification", "abuse", "integrity"}, false)
	add("passivity", "Passivity / stalling", "Passivity", "", GroupPenalties, "penalty", "penalty", "yellow", 0, []string{"foul", "stalling", "shido"}, false)
	add("false_attack", "False attack", "False attack", "", GroupPenalties, "penalty", "penalty", "yellow", 0, []string{"foul", "judo", "shido"}, false)
	add("illegal_technique", "Illegal technique (kinjite)", "Kinjite-waza", "禁じ手", GroupPenalties, "penalty", "penalty", "red", 0, []string{"foul", "prohibited", "hansoku"}, false)
	add("eye_gouge_foul", "Eye gouge (prohibited)", "Kyusho-me", "急所目", GroupPenalties, "penalty", "penalty", "red", 0, []string{"foul", "prohibited", "dangerous"}, false)
	add("dangerous_slam", "Dangerous slam", "Dangerous slam", "", GroupPenalties, "penalty", "penalty", "red", 0, []string{"foul", "prohibited", "dangerous"}, false)
	add("biting", "Biting", "Biting", "", GroupPenalties, "penalty", "penalty", "red", 0, []string{"foul", "prohibited", "abuse"}, false)
	add("headbutt_foul", "Headbutt (prohibited)", "Atama-uchi", "頭打ち", GroupPenalties, "penalty", "penalty", "red", 0, []string{"foul", "prohibited", "dangerous"}, false)
	add("heel_hook_foul", "Heel hook (prohibited)", "Heel hook", "", GroupPenalties, "penalty", "penalty", "red", 0, []string{"foul", "prohibited", "leg lock"}, false)

	return out
}

var (
	byIDOnce sync.Once
	byID     map[string]Technique

	groupsOnce sync.Once
	groupIndex map[string]Group
)

func buildIndexes() {
	byIDOnce.Do(func() {
		byID = make(map[string]Technique, len(catalogue))
		for _, tech := range catalogue {
			byID[tech.ID] = tech
		}
	})
	groupsOnce.Do(func() {
		groupIndex = make(map[string]Group, len(groupList))
		for _, group := range groupList {
			groupIndex[group.ID] = group
		}
	})
}

func copyTechnique(in Technique) Technique {
	if in.Tags != nil {
		in.Tags = append([]string(nil), in.Tags...)
	}
	return in
}

// All returns a copy of the catalogue in its deterministic source order.
func All() []Technique {
	out := make([]Technique, len(catalogue))
	for i, tech := range catalogue {
		out[i] = copyTechnique(tech)
	}
	return out
}

// ByID looks up a technique by its lowercase snake_case identifier.
func ByID(id string) (Technique, bool) {
	buildIndexes()
	tech, ok := byID[strings.ToLower(strings.TrimSpace(id))]
	if !ok {
		return Technique{}, false
	}
	return copyTechnique(tech), true
}

// Groups returns the stable, ordered list of catalogue sections.
func Groups() []Group {
	return append([]Group(nil), groupList...)
}

// Search matches a case-insensitive substring against English names, romaji,
// Japanese script, group id/label, point keys and tags. An empty query matches
// everything; group may be a group id or label, and "" means no filter.
// limit <= 0 returns every match. Source order is always preserved.
func Search(query, group string, limit int) []Technique {
	buildIndexes()
	q := strings.ToLower(strings.TrimSpace(query))
	filter := strings.ToLower(strings.TrimSpace(group))

	out := make([]Technique, 0, len(catalogue))
	for _, tech := range catalogue {
		if filter != "" && !groupMatches(tech.Group, filter) {
			continue
		}
		if q != "" && !matchesQuery(tech, q) {
			continue
		}
		out = append(out, copyTechnique(tech))
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func groupMatches(groupID, filter string) bool {
	group, ok := groupIndex[groupID]
	if !ok {
		return strings.EqualFold(groupID, filter)
	}
	return strings.EqualFold(group.ID, filter) || strings.Contains(strings.ToLower(group.Label), filter)
}

func matchesQuery(tech Technique, query string) bool {
	fields := []string{tech.English, tech.Romaji, tech.Japanese, tech.Group, tech.PointKey, tech.Kind, tech.Card}
	for _, field := range fields {
		if field != "" && strings.Contains(strings.ToLower(field), query) {
			return true
		}
	}
	if group, ok := groupIndex[tech.Group]; ok && strings.Contains(strings.ToLower(group.Label), query) {
		return true
	}
	for _, tag := range tech.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}
