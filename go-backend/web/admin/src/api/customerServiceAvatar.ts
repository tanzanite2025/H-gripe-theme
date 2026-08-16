import axios from '@/utils/axios'

export type CustomerServiceAvatarUserID = string | number

const avatarPath = (userID: CustomerServiceAvatarUserID): string => (
  `/api/admin/settings/public-chat-agents/${encodeURIComponent(String(userID))}/avatar`
)

export const customerServiceAvatarApi = {
  async upload(userID: CustomerServiceAvatarUserID, file: File): Promise<string> {
    const formData = new FormData()
    formData.append('file', file)

    const response = await axios.post(avatarPath(userID), formData)
    const avatar = String(response.data?.avatar || '').trim()
    if (!avatar) throw new Error('Avatar upload returned no URL')
    return avatar
  },

  async remove(userID: CustomerServiceAvatarUserID): Promise<void> {
    await axios.delete(avatarPath(userID))
  },
}

export default customerServiceAvatarApi
