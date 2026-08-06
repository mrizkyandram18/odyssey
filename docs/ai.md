# AI Integration (Future)

> **Status:** Future consideration. No AI features are planned for the MVP.
> This document exists to ensure AI-readiness in the domain model and to
> prevent feature drift. See [ADR-001](decisions/ADR-001-project-scope.md).

## Guiding Principle

If Odyssey ever integrates AI, it must serve the **cooperative adventure**
vision — never as a shortcut to replace human creativity or family
interaction. AI is an **amplifier** for the family's play, not a replacement
for it.

## Potential AI Roles

| Area | Concept | Description |
|---|---|---|
| **AI Story** | AI Dungeon Master | A generative narrative engine that adapts the story based on the family's choices and submissions. The AI does not decide for the family — it narrates consequences. |
| **AI Quest** | Quest Generation | Procedurally generate daily turns or short quests tailored to the family's interests and past submissions. Family members review and approve before playing. |
| **AI Quiz** | Educational Challenges | Generate custom quiz questions (history, science, pop culture) as puzzle challenges. Questions are explainable and verifiable. |
| **AI Comic** | Illustration Engine | Convert a family's written story submission into a multi-panel comic illustration. The family retains the original text and can edit the AI output. |
| **AI Judge** | Submission Evaluation | An AI that provides gentle, constructive feedback on creative submissions (e.g., "this story has a strong beginning; consider adding a twist"). Feedback is advisory, never a grade. |

## Non-Negotiables

- **Transparency:** If an AI generated or modified something, it is clearly
  labeled as such. No deepfakes, no impersonation of family members.
- **Human-in-the-loop:** AI outputs are always reviewed or can be edited by
  the family. The family's human input is the authoritative source.
- **No AI authentication.** Gatekeeper remains the sole auth path; AI is
  never an identity mechanism.
- **No AI for progression gating.** AI may suggest quests or generate
  content, but it never decides whether a player levels up or earns a Relic.
  Game rules stay deterministic.

## MVP Relevance

The `Submission.kind` discriminator (`STORY`, `COMIC`, `PHOTO`, `VIDEO`) is
designed with AI in mind: a future AI Comic feature would take a `STORY`
submission and produce a `COMIC` submission. The domain model already supports
this without change.

## Future Decision Points

When the time comes to evaluate AI integration, the following decisions will
be captured in new ADRs:

- **ADR-005: AI Provider Selection** — which model/provider to use.
- **ADR-006: AI Content Review Flow** — how human review is enforced.
- **ADR-007: AI Cost & Rate Limiting** — since AI inference has real cost,
  rate limits must be family-friendly and transparent.
