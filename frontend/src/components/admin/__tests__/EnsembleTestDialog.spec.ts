import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import EnsembleTestDialog from '../EnsembleTestDialog.vue'

describe('EnsembleTestDialog', () => {
  it('成员成本为 null 时显示未返回且不会调用 toFixed', () => {
    const wrapper = mount(EnsembleTestDialog, {
      props: {
        show: true,
        testing: false,
        error: '',
        result: null,
        events: [{
          type: 'member_finished',
          model: 'gpt-5',
          platform: 'openai',
          role: 'proposer',
          member: {
            model: 'gpt-5',
            platform: 'openai',
            role: 'proposer',
            status: 'failed',
            duration_ms: 20,
            cost: null,
            error: 'upstream failed'
          }
        } as any]
      },
      global: {
        stubs: {
          BaseDialog: { template: '<div><slot /><slot name="footer" /></div>' },
          Icon: true
        }
      }
    })

    expect(wrapper.text()).toContain('未返回')
    expect(wrapper.text()).toContain('upstream failed')
  })
})
