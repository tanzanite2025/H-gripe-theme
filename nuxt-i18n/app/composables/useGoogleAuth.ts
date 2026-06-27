/**
 * Google Identity Services (GSI) 封装 Composable
 * 用于处理 Google Sign-In 的初始化和登录流程
 */

import { ref, onMounted } from 'vue'
import { useRuntimeConfig } from '#imports'

// Google Identity Services 类型定义
declare global {
    interface Window {
        google?: {
            accounts: {
                id: {
                    initialize: (config: GoogleIdConfig) => void
                    prompt: (callback?: (notification: PromptNotification) => void) => void
                    renderButton: (element: HTMLElement, options: GoogleButtonOptions) => void
                    cancel: () => void
                    revoke: (email: string, callback: () => void) => void
                }
            }
        }
    }
}

interface GoogleIdConfig {
    client_id: string
    callback: (response: GoogleCredentialResponse) => void
    auto_select?: boolean
    cancel_on_tap_outside?: boolean
    context?: 'signin' | 'signup' | 'use'
    ux_mode?: 'popup' | 'redirect'
    login_uri?: string
    native_callback?: (response: GoogleCredentialResponse) => void
    itp_support?: boolean
}

interface GoogleCredentialResponse {
    credential: string // JWT ID Token
    select_by: string
    clientId?: string
}

interface PromptNotification {
    isNotDisplayed: () => boolean
    isSkippedMoment: () => boolean
    isDismissedMoment: () => boolean
    getNotDisplayedReason: () => string
    getSkippedReason: () => string
    getDismissedReason: () => string
}

interface GoogleButtonOptions {
    type?: 'standard' | 'icon'
    theme?: 'outline' | 'filled_blue' | 'filled_black'
    size?: 'large' | 'medium' | 'small'
    text?: 'signin_with' | 'signup_with' | 'continue_with' | 'signin'
    shape?: 'rectangular' | 'pill' | 'circle' | 'square'
    logo_alignment?: 'left' | 'center'
    width?: number
    locale?: string
}

export function useGoogleAuth() {
    const config = useRuntimeConfig()
    const clientId = config.public?.googleClientId as string || ''

    const isLoaded = ref(false)
    const isLoading = ref(false)
    const error = ref<string | null>(null)

    // 加载 Google Identity Services SDK
    const loadGsiScript = (): Promise<void> => {
        return new Promise((resolve, reject) => {
            // 已经加载
            if (window.google?.accounts?.id) {
                isLoaded.value = true
                resolve()
                return
            }

            // 检查是否已经有 script 标签
            const existingScript = document.querySelector('script[src*="accounts.google.com/gsi/client"]')
            if (existingScript) {
                existingScript.addEventListener('load', () => {
                    isLoaded.value = true
                    resolve()
                })
                return
            }

            // 创建新的 script 标签
            const script = document.createElement('script')
            script.src = 'https://accounts.google.com/gsi/client'
            script.async = true
            script.defer = true

            script.onload = () => {
                isLoaded.value = true
                resolve()
            }

            script.onerror = () => {
                error.value = 'Failed to load Google Sign-In SDK'
                reject(new Error('Failed to load Google Sign-In SDK'))
            }

            document.head.appendChild(script)
        })
    }

    // 初始化 Google Sign-In
    const initialize = async (callback: (response: GoogleCredentialResponse) => void): Promise<boolean> => {
        if (!clientId) {
            error.value = 'Google Client ID not configured'
            console.warn('[useGoogleAuth] Google Client ID not configured')
            return false
        }

        isLoading.value = true
        error.value = null

        try {
            await loadGsiScript()

            if (!window.google?.accounts?.id) {
                throw new Error('Google Identity Services not available')
            }

            window.google.accounts.id.initialize({
                client_id: clientId,
                callback,
                auto_select: false,
                cancel_on_tap_outside: true,
                context: 'signin',
                ux_mode: 'popup',
                itp_support: true
            })

            isLoading.value = false
            return true
        } catch (err) {
            error.value = err instanceof Error ? err.message : 'Google Sign-In initialization failed'
            isLoading.value = false
            return false
        }
    }

    // 显示 Google 登录弹窗
    const prompt = () => {
        if (!window.google?.accounts?.id) {
            error.value = 'Google Sign-In not initialized'
            return
        }

        window.google.accounts.id.prompt((notification) => {
            if (notification.isNotDisplayed()) {
                console.warn('[useGoogleAuth] Prompt not displayed:', notification.getNotDisplayedReason())
                // 可能是浏览器阻止了弹窗，提示用户手动点击
                error.value = 'Google 登录弹窗被阻止，请检查浏览器设置或稍后重试'
            }
        })
    }

    // 渲染 Google 登录按钮 (备用方案)
    const renderButton = (element: HTMLElement, options?: Partial<GoogleButtonOptions>) => {
        if (!window.google?.accounts?.id) {
            console.warn('[useGoogleAuth] Google Sign-In not initialized, cannot render button')
            return
        }

        window.google.accounts.id.renderButton(element, {
            type: 'standard',
            theme: 'filled_black',
            size: 'large',
            shape: 'pill',
            text: 'continue_with',
            ...options
        })
    }

    return {
        isLoaded,
        isLoading,
        error,
        clientId,
        initialize,
        prompt,
        renderButton,
        loadGsiScript
    }
}
