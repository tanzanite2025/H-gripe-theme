export const readBrowserCookie = (name: string) => {
  if (!import.meta.client || typeof document === 'undefined') {
    return ''
  }

  const prefix = `${encodeURIComponent(name)}=`
  const cookie = document.cookie
    .split(';')
    .map((item) => item.trim())
    .find((item) => item.startsWith(prefix))

  return cookie ? decodeURIComponent(cookie.slice(prefix.length)) : ''
}

export const hasBrowserCookie = (name: string) => readBrowserCookie(name).length > 0
