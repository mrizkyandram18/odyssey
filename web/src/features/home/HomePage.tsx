import { LinearPath } from '../stepper/LinearPath'
import { OnboardingModal } from './OnboardingModal'

export function HomePage() {
  return (
    <>
      <OnboardingModal />
      <LinearPath />
    </>
  )
}
