const CLASS_NAME = 'hide-sidebar-handles'
const activeTokens = new Set<string>()

const updateClass = () => {
  if (typeof document === 'undefined') return
  if (activeTokens.size > 0) {
    document.body.classList.add(CLASS_NAME)
  } else {
    document.body.classList.remove(CLASS_NAME)
  }
}

export const setSidebarHandlesHidden = (token: string, hidden: boolean) => {
  if (!token) return
  if (hidden) {
    activeTokens.add(token)
  } else {
    activeTokens.delete(token)
  }
  updateClass()
}
