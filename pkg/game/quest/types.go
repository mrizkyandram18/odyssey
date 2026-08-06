package quest

type QuestStatus string

const (
	QuestStatusPending QuestStatus = "PENDING"
	QuestStatusActive  QuestStatus = "ACTIVE"
	QuestStatusDone    QuestStatus = "DONE"
)

type ChallengeType string

const (
	ChallengeObservation ChallengeType = "OBSERVATION"
	ChallengeResearch    ChallengeType = "RESEARCH"
	ChallengePuzzle      ChallengeType = "PUZZLE"
	ChallengeMovement    ChallengeType = "MOVEMENT"
	ChallengeDraw        ChallengeType = "DRAW"
	ChallengeWrite       ChallengeType = "WRITE"
)

type ChallengeStatus string

const (
	ChallengeStatusPending ChallengeStatus = "PENDING"
	ChallengeStatusDone    ChallengeStatus = "DONE"
)

type QuestType string

const (
	QuestTypeSolo     QuestType = "SOLO"
	QuestTypeRelay    QuestType = "RELAY"
	QuestTypeGroup    QuestType = "GROUP"
	QuestTypeCreative QuestType = "CREATIVE"
)
