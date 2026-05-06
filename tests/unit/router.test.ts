import { describe, it, expect } from 'vitest'
import router from '@/router'

describe('Router Configuration', () => {
  it('has five routes defined (v0.2.0)', () => {
    // v0.2.0 新增 /projects 和 /history 路由
    expect(router.getRoutes()).toHaveLength(5)
  })

  it('has a home route at /', () => {
    const homeRoute = router.getRoutes().find(r => r.path === '/')
    expect(homeRoute).toBeDefined()
    expect(homeRoute?.name).toBe('home')
  })

  it('has a scan route at /scan', () => {
    const scanRoute = router.getRoutes().find(r => r.path === '/scan')
    expect(scanRoute).toBeDefined()
    expect(scanRoute?.name).toBe('scan')
  })

  it('has a settings route at /settings', () => {
    const settingsRoute = router.getRoutes().find(r => r.path === '/settings')
    expect(settingsRoute).toBeDefined()
    expect(settingsRoute?.name).toBe('settings')
  })

  it('has a projects route at /projects (v0.2.0)', () => {
    const projectsRoute = router.getRoutes().find(r => r.path === '/projects')
    expect(projectsRoute).toBeDefined()
    expect(projectsRoute?.name).toBe('projects')
  })

  it('has a history route at /history (v0.2.0)', () => {
    const historyRoute = router.getRoutes().find(r => r.path === '/history')
    expect(historyRoute).toBeDefined()
    expect(historyRoute?.name).toBe('history')
  })

  it('home route uses lazy-loaded component', () => {
    const homeRoute = router.getRoutes().find(r => r.path === '/')
    // Lazy loaded routes have a component function
    expect(typeof homeRoute?.components?.default).toBe('function')
  })

  it('scan route uses lazy-loaded component', () => {
    const scanRoute = router.getRoutes().find(r => r.path === '/scan')
    expect(typeof scanRoute?.components?.default).toBe('function')
  })

  it('settings route uses lazy-loaded component', () => {
    const settingsRoute = router.getRoutes().find(r => r.path === '/settings')
    expect(typeof settingsRoute?.components?.default).toBe('function')
  })

  it('projects route uses lazy-loaded component (v0.2.0)', () => {
    const projectsRoute = router.getRoutes().find(r => r.path === '/projects')
    expect(typeof projectsRoute?.components?.default).toBe('function')
  })

  it('history route uses lazy-loaded component (v0.2.0)', () => {
    const historyRoute = router.getRoutes().find(r => r.path === '/history')
    expect(typeof historyRoute?.components?.default).toBe('function')
  })

  it('uses web history mode', () => {
    // Web history mode uses HTML5 pushState
    expect(router.options.history).toBeDefined()
  })
})
