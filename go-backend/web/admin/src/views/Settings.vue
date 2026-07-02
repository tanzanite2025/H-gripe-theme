<template>
  <div class="settings-page">
    <div class="page-header">
      <h2>系统设置</h2>
      <el-button
        v-if="hasPermission('settings:edit') && activeTab !== 'public_chat'"
        type="primary"
        :loading="saving"
        @click="saveSettings"
      >
        保存设置
      </el-button>
    </div>

    <el-card>
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <!-- 站点设置 -->
        <el-tab-pane label="站点设置" name="site">
          <el-form
            ref="siteFormRef"
            :model="siteSettings"
            label-width="150px"
            style="max-width: 800px"
          >
            <el-form-item label="站点名称">
              <el-input v-model="siteSettings.site_name" />
            </el-form-item>

            <el-form-item label="站点描述">
              <el-input v-model="siteSettings.site_description" type="textarea" :rows="3" />
            </el-form-item>

            <el-form-item label="站点 Logo">
              <el-input v-model="siteSettings.site_logo" placeholder="Logo URL" />
            </el-form-item>

            <el-form-item label="联系邮箱">
              <el-input v-model="siteSettings.contact_email" />
            </el-form-item>

            <el-form-item label="联系电话">
              <el-input v-model="siteSettings.contact_phone" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 邮件设置 -->
        <el-tab-pane label="邮件设置" name="email">
          <el-form
            ref="emailFormRef"
            :model="emailSettings"
            label-width="150px"
            style="max-width: 800px"
          >
            <el-form-item label="SMTP 主机">
              <el-input v-model="emailSettings.smtp_host" />
            </el-form-item>

            <el-form-item label="SMTP 端口">
              <el-input-number v-model="emailSettings.smtp_port" :min="1" :max="65535" />
            </el-form-item>

            <el-form-item label="SMTP 用户名">
              <el-input v-model="emailSettings.smtp_username" />
            </el-form-item>

            <el-form-item label="SMTP 密码">
              <el-input v-model="emailSettings.smtp_password" type="password" show-password />
            </el-form-item>

            <el-form-item label="发件人邮箱">
              <el-input v-model="emailSettings.from_email" />
            </el-form-item>

            <el-form-item label="发件人名称">
              <el-input v-model="emailSettings.from_name" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- SEO 设置 -->
        <el-tab-pane label="SEO 设置" name="seo">
          <el-form
            ref="seoFormRef"
            :model="seoSettings"
            label-width="150px"
            style="max-width: 800px"
          >
            <el-form-item label="Meta 标题">
              <el-input v-model="seoSettings.meta_title" />
            </el-form-item>

            <el-form-item label="Meta 描述">
              <el-input v-model="seoSettings.meta_description" type="textarea" :rows="3" />
            </el-form-item>

            <el-form-item label="Meta 关键词">
              <el-input v-model="seoSettings.meta_keywords" placeholder="用逗号分隔" />
            </el-form-item>

            <el-form-item label="Google Analytics">
              <el-input v-model="seoSettings.google_analytics" placeholder="GA 跟踪 ID" />
            </el-form-item>

            <el-form-item label="Google Tag Manager">
              <el-input v-model="seoSettings.google_tag_manager" placeholder="GTM ID" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 社交媒体设置 -->
        <el-tab-pane label="社交媒体" name="social">
          <el-form
            ref="socialFormRef"
            :model="socialSettings"
            label-width="150px"
            style="max-width: 800px"
          >
            <el-form-item label="Facebook">
              <el-input v-model="socialSettings.facebook" placeholder="Facebook 页面 URL" />
            </el-form-item>

            <el-form-item label="Twitter">
              <el-input v-model="socialSettings.twitter" placeholder="Twitter 账号 URL" />
            </el-form-item>

            <el-form-item label="Instagram">
              <el-input v-model="socialSettings.instagram" placeholder="Instagram 账号 URL" />
            </el-form-item>

            <el-form-item label="LinkedIn">
              <el-input v-model="socialSettings.linkedin" placeholder="LinkedIn 页面 URL" />
            </el-form-item>

            <el-form-item label="YouTube">
              <el-input v-model="socialSettings.youtube" placeholder="YouTube 频道 URL" />
            </el-form-item>

            <el-form-item label="微信">
              <el-input v-model="socialSettings.wechat" placeholder="微信二维码 URL" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- 支付设置 -->
        <el-tab-pane label="支付设置" name="payment">
          <el-form
            ref="paymentFormRef"
            :model="paymentSettings"
            label-width="150px"
            style="max-width: 800px"
          >
            <el-form-item label="支付网关">
              <el-select v-model="paymentSettings.gateway" placeholder="请选择支付网关">
                <el-option label="Stripe" value="stripe" />
                <el-option label="PayPal" value="paypal" />
                <el-option label="支付宝" value="alipay" />
                <el-option label="微信支付" value="wechat" />
              </el-select>
            </el-form-item>

            <el-form-item label="API Key">
              <el-input v-model="paymentSettings.api_key" type="password" show-password />
            </el-form-item>

            <el-form-item label="API Secret">
              <el-input v-model="paymentSettings.api_secret" type="password" show-password />
            </el-form-item>

            <el-form-item label="测试模式">
              <el-switch v-model="paymentSettings.test_mode" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Public Chat 客服 -->
        <el-tab-pane label="Public Chat 客服" name="public_chat">
          <div class="agent-panel">
            <el-alert
              title="Public Chat 只读取 Go customer_service_agent_profiles；公开展示需绑定 active 用户且角色为 admin、manager 或 support。"
              type="info"
              show-icon
              :closable="false"
            />

            <div class="panel-actions">
              <el-button
                type="primary"
                :loading="loadingPublicChatAgents"
                @click="fetchPublicChatAgents"
              >
                刷新客服概览
              </el-button>
            </div>

            <el-row :gutter="16" class="summary-row">
              <el-col :xs="24" :sm="8">
                <el-card shadow="never">
                  <div class="summary-label">Profile 总数</div>
                  <div class="summary-value">{{ publicChatAgentsSummary.profile_count || 0 }}</div>
                </el-card>
              </el-col>
              <el-col :xs="24" :sm="8">
                <el-card shadow="never">
                  <div class="summary-label">公开客服数</div>
                  <div class="summary-value">{{ publicChatAgentsSummary.exposed_agents || 0 }}</div>
                </el-card>
              </el-col>
            </el-row>

            <el-alert
              v-for="warning in publicChatAgentWarnings"
              :key="warning"
              class="agent-warning"
              :title="warning"
              type="warning"
              show-icon
              :closable="false"
            />

            <h3>Public chat agents</h3>
            <el-table
              v-loading="loadingPublicChatAgents"
              :data="publicChatAgents"
              border
              empty-text="暂无 public chat 客服 profile"
            >
              <el-table-column prop="id" label="Profile ID" width="100" />
              <el-table-column prop="agent_id" label="Agent ID" width="120" />
              <el-table-column prop="user_id" label="User ID" width="90" />
              <el-table-column prop="display_name" label="Name" min-width="140" />
              <el-table-column prop="email" label="Email" min-width="180" />
              <el-table-column prop="whatsapp" label="WhatsApp" width="140" />
              <el-table-column prop="raw_role" label="Raw role" width="150" />
              <el-table-column prop="normalized_role" label="Go role" width="120" />
              <el-table-column prop="user_status" label="User status" width="120" />
              <el-table-column prop="profile_status" label="Profile status" width="130" />
              <el-table-column prop="online_status" label="Online" width="100" />
              <el-table-column label="Public exposed" width="130">
                <template #default="{ row }">
                  <el-tag :type="row.exposed ? 'success' : 'danger'">
                    {{ row.exposed ? 'Yes' : 'No' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const activeTab = ref('site')
const saving = ref(false)

const siteSettings = reactive({
  site_name: '',
  site_description: '',
  site_logo: '',
  contact_email: '',
  contact_phone: ''
})

const emailSettings = reactive({
  smtp_host: '',
  smtp_port: 587,
  smtp_username: '',
  smtp_password: '',
  from_email: '',
  from_name: ''
})

const seoSettings = reactive({
  meta_title: '',
  meta_description: '',
  meta_keywords: '',
  google_analytics: '',
  google_tag_manager: ''
})

const socialSettings = reactive({
  facebook: '',
  twitter: '',
  instagram: '',
  linkedin: '',
  youtube: '',
  wechat: ''
})

const paymentSettings = reactive({
  gateway: '',
  api_key: '',
  api_secret: '',
  test_mode: true
})

const loadingPublicChatAgents = ref(false)
const publicChatAgentsOverview = ref(null)
const publicChatAgentsSummary = computed(() => publicChatAgentsOverview.value?.summary || {})
const publicChatAgents = computed(() => publicChatAgentsOverview.value?.agents || [])
const publicChatAgentWarnings = computed(() => publicChatAgentsOverview.value?.warnings || [])

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const fetchSettings = async (group) => {
  try {
    const response = await axios.get(`/api/admin/settings/${group}`)
    const settings = response.data.settings

    const targetSettings = {
      site: siteSettings,
      email: emailSettings,
      seo: seoSettings,
      social: socialSettings,
      payment: paymentSettings
    }[group]

    if (targetSettings && settings) {
      settings.forEach(setting => {
        const key = setting.key.replace(`${group}_`, '')
        if (key in targetSettings) {
          targetSettings[key] = setting.value
        }
      })
    }
  } catch (error) {
    console.error(`获取${group}设置失败`, error)
    ElMessage.error(`获取${group === 'site' ? '站点' : group}设置失败`)
  }
}

const handleTabChange = (tabName) => {
  if (tabName === 'public_chat') {
    fetchPublicChatAgents()
    return
  }
  fetchSettings(tabName)
}

const fetchPublicChatAgents = async () => {
  loadingPublicChatAgents.value = true
  try {
    const response = await axios.get('/api/admin/settings/public-chat-agents')
    publicChatAgentsOverview.value = response.data
  } catch (error) {
    console.error('获取 Public Chat 客服失败', error)
    ElMessage.error('获取 Public Chat 客服失败')
  } finally {
    loadingPublicChatAgents.value = false
  }
}

const saveSettings = async () => {
  const currentSettings = {
    site: siteSettings,
    email: emailSettings,
    seo: seoSettings,
    social: socialSettings,
    payment: paymentSettings
  }[activeTab.value]
  if (!currentSettings) return

  const settingsArray = Object.entries(currentSettings).map(([key, value]) => ({
    key: `${activeTab.value}_${key}`,
    value: String(value),
    group: activeTab.value,
    locale: 'en',
    is_public: activeTab.value !== 'email' && activeTab.value !== 'payment'
  }))

  saving.value = true
  try {
    await axios.post('/api/admin/settings/batch', {
      settings: settingsArray
    })
    ElMessage.success('设置保存成功')
  } catch (error) {
    ElMessage.error('设置保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchSettings('site')
})
</script>

<style scoped>
.settings-page {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 24px;
  color: #303133;
}

.agent-panel {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.panel-actions {
  display: flex;
  justify-content: flex-end;
}

.summary-row {
  width: 100%;
}

.summary-label {
  color: #606266;
  font-size: 13px;
  margin-bottom: 10px;
}

.summary-value {
  font-size: 28px;
  font-weight: 600;
  color: #303133;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.agent-warning {
  margin-top: -6px;
}

.sql-block {
  margin: 0;
  padding: 12px;
  white-space: pre-wrap;
  word-break: break-word;
  background: #f5f7fa;
  border-radius: 6px;
  color: #303133;
}
</style>
