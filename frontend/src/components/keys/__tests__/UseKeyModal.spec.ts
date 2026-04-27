import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { nextTick } from 'vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => key
  })
}))

vi.mock('@/composables/useClipboard', () => ({
  useClipboard: () => ({
    copyToClipboard: vi.fn().mockResolvedValue(true)
  })
}))

import UseKeyModal from '../UseKeyModal.vue'

function mountUseKeyModal() {
  return mount(UseKeyModal, {
    props: {
      show: true,
      apiKey: 'sk-test',
      baseUrl: 'https://example.com/v1',
      platform: 'openai'
    },
    global: {
      stubs: {
        BaseDialog: {
          template: '<div><slot /><slot name="footer" /></div>'
        },
        Icon: {
          template: '<span />'
        }
      }
    }
  })
}

describe('UseKeyModal', () => {
  it('renders GPT-5.4 mini entry in OpenCode config', async () => {
    const wrapper = mountUseKeyModal()

    const opencodeTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.opencode')
    )

    expect(opencodeTab).toBeDefined()
    await opencodeTab!.trigger('click')
    await nextTick()

    const codeBlock = wrapper.find('pre code')
    expect(codeBlock.exists()).toBe(true)
    expect(codeBlock.text()).toContain('"name": "GPT-5.4 Mini"')
    expect(codeBlock.text()).not.toContain('"name": "GPT-5.4 Nano"')
  })

  it('uses GPT-5.5 as the default Codex CLI model', async () => {
    const wrapper = mountUseKeyModal()

    let codeBlock = wrapper.find('pre code')
    expect(codeBlock.text()).toContain('model = "gpt-5.5"')
    expect(codeBlock.text()).toContain('review_model = "gpt-5.5"')
    expect(codeBlock.text()).not.toContain('model = "gpt-5.4"')

    const codexWsTab = wrapper.findAll('button').find((button) =>
      button.text().includes('keys.useKeyModal.cliTabs.codexCliWs')
    )

    expect(codexWsTab).toBeDefined()
    await codexWsTab!.trigger('click')
    await nextTick()

    codeBlock = wrapper.find('pre code')
    expect(codeBlock.text()).toContain('model = "gpt-5.5"')
    expect(codeBlock.text()).toContain('review_model = "gpt-5.5"')
    expect(codeBlock.text()).not.toContain('model = "gpt-5.4"')
  })
})
