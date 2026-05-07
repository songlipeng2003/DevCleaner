import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  ProjectScanResult,
  CleanableItem,
  ProjectCleanConfig,
  CleanResult,
} from '@/types'
import * as tauriApi from '@/services/tauri'

export const useProjectCleanStore = defineStore('projectClean', () => {
  // ============== State ==============
  const projects = ref<ProjectScanResult[]>([])
  const isScanning = ref(false)
  const isCleaning = ref(false)
  const error = ref<string | null>(null)
  const selectedItems = ref<CleanableItem[]>([])

  // Default configuration
  const config = ref<ProjectCleanConfig>({
    scanPaths: ['~/projects'],
    includeTypes: ['node', 'rust', 'python', 'go', 'java', 'dotnet', 'ruby', 'unity'],
    excludePatterns: [],
    maxDepth: 3,
    minSize: 10 * 1024 * 1024, // 10MB
  })

  // ============== Getters ==============

  // Get all cleanable items from all projects
  const allCleanableItems = computed(() => {
    return projects.value.flatMap(project =>
      project.cleanableItems.filter(item => item.cleanable)
    )
  })

  // Total size of all cleanable items
  const totalCleanableSize = computed(() => {
    return allCleanableItems.value.reduce((sum, item) => sum + item.size, 0)
  })

  // Total count of cleanable items
  const totalCleanableItems = computed(() => {
    return allCleanableItems.value.length
  })

  // Total size of selected items
  const selectedSize = computed(() => {
    return selectedItems.value.reduce((sum, item) => sum + item.size, 0)
  })

  // Selected items count
  const selectedCount = computed(() => {
    return selectedItems.value.length
  })

  // Filter projects by type
  const projectsByType = computed(() => {
    const grouped: Record<string, ProjectScanResult[]> = {}
    for (const project of projects.value) {
      if (!grouped[project.type]) {
        grouped[project.type] = []
      }
      grouped[project.type].push(project)
    }
    return grouped
  })

  // Get projects sorted by total size
  const sortedProjects = computed(() => {
    return [...projects.value].sort((a, b) => b.size - a.size)
  })

  // Check if an item is selected
  const isSelected = (itemId: string) => {
    return selectedItems.value.some(item => item.id === itemId)
  }

  // ============== Actions ==============

  // Scan projects for cleanable directories
  async function scanProjects(): Promise<ProjectScanResult[]> {
    isScanning.value = true
    error.value = null

    try {
      const results = await tauriApi.scanProjects(
        config.value.scanPaths,
        config.value.maxDepth
      )
      projects.value = results

      // Clear selection when scanning new results
      selectedItems.value = []

      return results
    } catch (e) {
      error.value = e instanceof Error ? e.message : '扫描项目失败'
      throw e
    } finally {
      isScanning.value = false
    }
  }

  // Clean selected items
  async function cleanSelected(): Promise<CleanResult> {
    if (selectedItems.value.length === 0) {
      throw new Error('没有选中的清理项')
    }

    isCleaning.value = true
    error.value = null

    const paths = selectedItems.value.map(item => item.path)

    try {
      const result = await tauriApi.cleanPaths(paths)

      // Remove cleaned items from projects
      const cleanedIds = new Set(selectedItems.value.map(i => i.id))
      for (const project of projects.value) {
        project.cleanableItems = project.cleanableItems.filter(
          item => !cleanedIds.has(item.id)
        )
        // Recalculate project size
        project.size = project.cleanableItems.reduce((sum, item) => sum + item.size, 0)
      }

      // Remove projects with no cleanable items
      projects.value = projects.value.filter(p => p.cleanableItems.length > 0)

      // Record to history
      if (result.cleaned > 0) {
        await tauriApi.recordCleanHistory(
          'project',
          '项目清理',
          result.cleaned,
          result.file_num,
          paths
        )
      }

      // Clear selection
      selectedItems.value = []

      return result
    } catch (e) {
      error.value = e instanceof Error ? e.message : '清理失败'
      throw e
    } finally {
      isCleaning.value = false
    }
  }

  // Clean specific items (not from selection)
  async function cleanItems(items: CleanableItem[]): Promise<CleanResult> {
    if (items.length === 0) {
      throw new Error('没有要清理的项目')
    }

    isCleaning.value = true
    error.value = null

    const paths = items.map(item => item.path)

    try {
      const result = await tauriApi.cleanPaths(paths)

      // Remove cleaned items from projects
      const cleanedIds = new Set(items.map(i => i.id))
      for (const project of projects.value) {
        project.cleanableItems = project.cleanableItems.filter(
          item => !cleanedIds.has(item.id)
        )
        project.size = project.cleanableItems.reduce((sum, item) => sum + item.size, 0)
      }

      projects.value = projects.value.filter(p => p.cleanableItems.length > 0)

      // Record to history
      if (result.cleaned > 0) {
        await tauriApi.recordCleanHistory(
          'project',
          '项目清理',
          result.cleaned,
          result.file_num,
          paths
        )
      }

      return result
    } catch (e) {
      error.value = e instanceof Error ? e.message : '清理失败'
      throw e
    } finally {
      isCleaning.value = false
    }
  }

  // Toggle item selection
  function toggleSelection(item: CleanableItem) {
    const index = selectedItems.value.findIndex(i => i.id === item.id)
    if (index >= 0) {
      selectedItems.value.splice(index, 1)
    } else {
      selectedItems.value.push(item)
    }
  }

  // Select all cleanable items
  function selectAll() {
    selectedItems.value = [...allCleanableItems.value]
  }

  // Clear all selections
  function clearSelection() {
    selectedItems.value = []
  }

  // Select items by type
  function selectByType(type: CleanableItem['type']) {
    const items = allCleanableItems.value.filter(item => item.type === type)
    for (const item of items) {
      if (!isSelected(item.id)) {
        selectedItems.value.push(item)
      }
    }
  }

  // Deselect items by type
  function deselectByType(type: CleanableItem['type']) {
    selectedItems.value = selectedItems.value.filter(item => item.type !== type)
  }

  // Select safe items only (riskLevel: safe)
  function selectSafeOnly() {
    selectedItems.value = allCleanableItems.value.filter(item => item.cleanable)
  }

  // Update scan paths
  function setScanPaths(paths: string[]) {
    config.value.scanPaths = paths
  }

  // Update configuration
  function setConfig(newConfig: Partial<ProjectCleanConfig>) {
    config.value = { ...config.value, ...newConfig }
  }

  // Get projects filtered by criteria
  function getFilteredProjects(criteria: {
    type?: string
    minSize?: number
    riskLevel?: 'safe' | 'moderate' | 'careful'
  }): ProjectScanResult[] {
    let filtered = [...projects.value]

    if (criteria.type) {
      filtered = filtered.filter(p => p.type === criteria.type)
    }

    if (criteria.minSize) {
      filtered = filtered.filter(p => p.size >= criteria.minSize)
    }

    if (criteria.riskLevel) {
      filtered = filtered.filter(p => p.riskLevel === criteria.riskLevel)
    }

    return filtered
  }

  // Format bytes to human readable
  function formatSize(bytes: number): string {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // Open path in file manager
  async function openPath(path: string) {
    return tauriApi.openPath(path)
  }

  // Clear all projects
  function clearProjects() {
    projects.value = []
    selectedItems.value = []
    error.value = null
  }

  // ============== Return ==============
  return {
    // State
    projects,
    isScanning,
    isCleaning,
    error,
    selectedItems,
    config,

    // Getters
    allCleanableItems,
    totalCleanableSize,
    totalCleanableItems,
    selectedSize,
    selectedCount,
    projectsByType,
    sortedProjects,
    isSelected,

    // Actions
    scanProjects,
    cleanSelected,
    cleanItems,
    toggleSelection,
    selectAll,
    clearSelection,
    selectByType,
    deselectByType,
    selectSafeOnly,
    setScanPaths,
    setConfig,
    getFilteredProjects,
    formatSize,
    openPath,
    clearProjects,
  }
})
