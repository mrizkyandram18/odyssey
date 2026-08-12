# Glossary

## A

**Authentication Provider** — The abstraction through which Odyssey verifies
user identity. Odyssey defines an `Authenticator` port in its domain layer;
an adapter (e.g., `FirestoreAuthenticator`) implements it by reading
Gatekeeper's device documents. The domain never references Firestore directly.

## B

**BOTH Login Mode** — A Gatekeeper authentication mode requiring both device
trust compliance (online device, correct build, required permissions) and a
password credential. Odyssey reuses this as its sole authentication path.

**Branch** — A divergent path in a quest that leads to a different story
outcome or challenge.

**Builder** — A narrative role focused on creation (drawing, building,
composing). No mechanical stat bonus or advantage.

## C

**Exercise** — A short, objective within a quest (e.g., "find three red
flowers," "solve this riddle"). Exercises may require collaboration.

**Gift** — A reward container earned through quest completion or Explorer
Level-ups. MVP Gifts have known, fixed contents (no randomization / no
gambling). Contains Collections, cosmetics, or creative-tool unlocks.

**Comic** — A multi-panel illustrated story submission within a Creative Mission.

**Creative Mission** — A quest type where the family builds, draws, or tells
stories together. Accepts Story, Comic, Photo, and Video submissions.

**Creative Space** — A sandbox area in Odyssey where family members can build,
decorate, or compose collaboratively (e.g., a shared garden, a story canvas).

**Family** — The family group as a single party. All progress is tracked at both
the individual (Explorer) and crew level.

## E

**Explorer** — A family member's representation in the shared world. An
Explorer has a name, a role, an Explorer Level, and a collection of Collections.
(Not called "Avatar" to avoid implying purely cosmetic progression.)

**Explorer Level** — The individual progression level, advancing through XP.

**Inspiration** — A soft currency earned through creative-space contributions
and peer reactions. Spent on advanced creative tools and cosmetics.

## M

**MVP** — Minimum Viable Product. The first phase of Odyssey, scoped to the
smallest complete cooperative loop that delivers a real "aha" moment.

**Milestone** — A significant individual or group accomplishment that awards
XP, a Collection, or unlocks new content.

## P

**Party** — See **Family**.

**Photo** — A real-world photograph submission within a Creative Mission.

**Progress Sync** — *(Future.)* The process of reconciling local browser state
with the server when connectivity is restored. Not in the MVP.

**Prompt** — A creative writing prompt or theme that spawns Story Submissions.

## Q

**Mission** — The primary unit of cooperative play. A quest is a sequence of
exercises threaded into a story. Missions can be Solo, Relay, Group, or
Creative (see Gameplay).

**Mission Completion** — When all exercises in a Mission are resolved, the family
receives XP, a Gift, and a story unlock.

## R

**Journey** — A themed area of the shared world (e.g., "The Whispering Woods,"
"The Clockwork City"). Journeys contain missions and creative spaces.

**Journey Progress** — A shared progress bar within a journey, filled by
completing missions and exercises. Milestones unlock new missions, Journey Gifts,
and the next journey.

**Collection** — A collectible item earned through quest completion, milestones, or
story fragments. Displayed in the crew gallery. Not currency — collected, not
spent. Tradable within the family (Phase 2).

**Reward** — A real-world reward issued by Family Reward. Odyssey never creates
or tracks rewards. See [ADR-004](decisions/ADR-004-reward-integration.md).

**Ride** — *(Reserved.)* A future mechanic for a shared, sequential experience
where each family member contributes a step. Not in MVP.

**Role** — A narrative framing for an Explorer (Seeker, Builder, Guide). No
mechanical advantage.

## S

**Seeker** — A narrative role focused on discovery and observation. No
mechanical stat bonus or advantage.

**Session** — An HMAC-signed, short-lived identity token issued by the Odyssey
backend after successful Gatekeeper BOTH login. Stored in browser local
storage and sent via the `Authorization: Bearer` header.

**Story** — A written narrative snippet submitted as part of a Creative Mission.

**Story Fragment** — A collectible found during missions. Reassembling a set
unlocks concept and narrative content.

**Submission** — A player's response to a challenge or creative prompt.
Submission kinds in Creative Missions: Story, Comic, Photo, Video.

## U

**UID** — The shared user identifier. The same UID is used across Gatekeeper,
Family Reward, and Odyssey to identify a single family member.

## V

**Video** — A short (≤ 30 s) video submission within a Creative Mission.

**Vision Document** — See [vision.md](vision.md).

## W

**World State** — The persisted, shared game state of the family's adventure:
completed missions, crew milestones, journey unlocks, and shared narrative
progress.
