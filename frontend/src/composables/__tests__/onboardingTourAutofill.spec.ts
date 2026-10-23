import { describe, expect, it } from 'vitest'

/**
 * Lightweight DOM autofill behavior used by the onboarding tour.
 * Mirrors useOnboardingTour.fillTourInputIfEmpty so required Create steps
 * cannot soft-lock on empty HTML5 validation.
 */
function fillTourInputIfEmpty(selector: string, value: string): boolean {
  const el = document.querySelector(selector) as HTMLInputElement | HTMLTextAreaElement | null
  if (!el) return false
  if (el.value.trim()) return true

  const prototype =
    el instanceof HTMLTextAreaElement
      ? window.HTMLTextAreaElement.prototype
      : window.HTMLInputElement.prototype
  const valueSetter = Object.getOwnPropertyDescriptor(prototype, 'value')?.set
  if (valueSetter) {
    valueSetter.call(el, value)
  } else {
    el.value = value
  }
  el.dispatchEvent(new Event('input', { bubbles: true }))
  el.dispatchEvent(new Event('change', { bubbles: true }))
  return Boolean(el.value.trim())
}

describe('onboarding tour required-field autofill', () => {
  it('fills an empty group name input and emits input for v-model sync', () => {
    document.body.innerHTML =
      '<input data-tour="group-form-name" type="text" required value="" />'
    const input = document.querySelector(
      '[data-tour="group-form-name"]'
    ) as HTMLInputElement
    let inputEvents = 0
    input.addEventListener('input', () => {
      inputEvents += 1
    })

    expect(fillTourInputIfEmpty('[data-tour="group-form-name"]', '测试分组')).toBe(true)
    expect(input.value).toBe('测试分组')
    expect(inputEvents).toBe(1)
  })

  it('does not overwrite a user-provided group name', () => {
    document.body.innerHTML =
      '<input data-tour="group-form-name" type="text" value="研发团队" />'
    const input = document.querySelector(
      '[data-tour="group-form-name"]'
    ) as HTMLInputElement

    expect(fillTourInputIfEmpty('[data-tour="group-form-name"]', '测试分组')).toBe(true)
    expect(input.value).toBe('研发团队')
  })
})
