# RCA: Quest Identifier Inconsistency

## Description
There is a design drift and type inconsistency regarding the `quest_id` identifier between the core domain schema, feature tables, and the Go backend models.

## Evidence
- **Core Schema (`001_initial_schema.sql`)**: 
  `odyssey_quests.id` is defined as `BIGINT`.
- **Feature Schema (`013_crew_interaction.sql`)**: 
  `odyssey_reactions.quest_id` is defined as `UUID`.
- **Go Domain Model (`pkg/game/domain.go`)**: 
  `Quest.ID` is mapped as `int64`.
  `Reaction.QuestID` is mapped as `*string`.
- **API (`api/reactions/index.go`)**: 
  Expects `quest_id` as a string (`*string` in JSON request).

## Impact
- You cannot insert canonical quest IDs (e.g., `101`) into `odyssey_reactions.quest_id` via SQL because PostgreSQL rejects the type cast to `UUID`.
- A direct SQL hotfix (`ALTER COLUMN quest_id TYPE BIGINT`) will break the Go backend, causing `Scan error` when the DB driver attempts to map a `BIGINT` to `*string`.

## Recommended Action (Refactor)
This must be resolved in a single, coordinated PR encompassing the entire stack:
1. **Decide on Canonical Design**: Choose whether Quests should definitively use `BIGINT` (Sequential ID) or `UUID` across the entire platform.
2. **Database Migration**: Write a proper migration to align `odyssey_reactions.quest_id` and any other referencing tables.
3. **Go Models**: Update `Reaction.QuestID`, `Challenge.QuestID`, etc., to use the chosen type (`int64` or `uuid.UUID` / `string`).
4. **API & Repositories**: Update all endpoints and DB scanning logic to handle the canonical type.
5. **Frontend**: Update the TypeScript interfaces (`QuestId: number | string`).
