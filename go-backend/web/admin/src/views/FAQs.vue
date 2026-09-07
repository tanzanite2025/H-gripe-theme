<template>
  <div class="flex h-full min-h-0 flex-col gap-3 overflow-hidden">
    <AdminPageHeader class="shrink-0" title="FAQ 管理" description="维护常见问题、页面归属、发布状态和展示顺序">
      <template #actions>
        <Button v-if="hasPermission('faq:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          添加 FAQ
        </Button>
      </template>
    </AdminPageHeader>

    <div class="shrink-0">
      <FAQFilterPanel
        :filters="filters"
        :page-filter-options="pageFilterOptions"
        :status-filter-options="statusFilterOptions"
        @apply="applyFilters"
        @reset="resetFilters"
      />
    </div>

    <FAQAccordionList
      class="min-h-0 flex-1"
      :loading="loading"
      :structure-loading="structureLoading"
      :faq-groups="faqGroups"
      :selected-faqs="selectedFAQs"
      :pagination="pagination"
      :structure-locales="structureLocales"
      :active-structure-locale="activeStructureLocale"
      :has-permission="hasPermission"
      :is-selected="isSelected"
      :plain-text="plainTextFromHTML"
      :status-tone="statusTone"
      :status-name="statusName"
      :visibility-name="visibilityName"
      :visibility-tone="visibilityTone"
      :domain-name="domainName"
      @switch-locale="switchStructureLocale"
      @toggle-faq="toggleFAQ"
      @edit="showEditDialog"
      @delete="requestDelete"
      @batch-delete="requestBatchDelete"
      @edit-page="showPageDialog"
      @create-faq="openCreateFAQDialog"
    />

    <FAQEditorDialog
      v-model:open="dialogVisible"
      :dialog-mode="dialogMode"
      :faq-form="faqForm"
      :form-errors="formErrors"
      :submitting="submitting"
      :faq-page-options="faqPageOptions"
      :language-options="languageOptions"
      :placement-locked="placementLocked"
      @submit="submitForm"
      @clear-error="clearFieldError"
      @update-answer="updateFAQAnswer"
    />

    <FAQPageEditorDialog
      v-model:open="pageDialogVisible"
      :page-form="pageForm"
      :submitting="pageSubmitting"
      :locale-name="localeName"
      @submit="submitPageForm"
    />

    <AdminConfirmDialog
      v-model:open="confirmation.open"
      :title="confirmation.title"
      :description="confirmation.description"
      :confirm-label="confirmation.confirmLabel"
      destructive
      @confirm="executeConfirmedAction"
    />
  </div>
</template>

<script setup lang="ts">
import { Plus } from '@lucide/vue'
import type { FAQStructurePage } from '@/lib/faqAdminPresentation'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import FAQAccordionList from '@/components/admin/faq/FAQAccordionList.vue'
import FAQEditorDialog from '@/components/admin/faq/FAQEditorDialog.vue'
import FAQFilterPanel from '@/components/admin/faq/FAQFilterPanel.vue'
import FAQPageEditorDialog from '@/components/admin/faq/FAQPageEditorDialog.vue'
import { Button } from '@/components/ui/button'
import { useFaqAdmin } from '@/composables/useFaqAdmin'

const {
  loading,
  structureLoading,
  faqGroups,
  activeStructureLocale,
  selectedFAQs,
  dialogVisible,
  dialogMode,
  submitting,
  placementLocked,
  pageDialogVisible,
  pageSubmitting,
  formErrors,
  filters,
  pagination,
  faqForm,
  pageForm,
  confirmation,
  statusFilterOptions,
  structureLocales,
  languageOptions,
  structurePageOptions,
  faqPageOptions,
  pageFilterOptions,
  hasPermission,
  localeName,
  statusName,
  statusTone,
  visibilityName,
  visibilityTone,
  domainName,
  plainTextFromHTML,
  clearFieldError,
  updateFAQAnswer,
  switchStructureLocale,
  applyFilters,
  resetFilters,
  showCreateDialog,
  showEditDialog,
  submitForm,
  showPageDialog,
  submitPageForm,
  isSelected,
  toggleFAQ,
  requestDelete,
  requestBatchDelete,
  executeConfirmedAction
} = useFaqAdmin()

const openCreateFAQDialog = (page: FAQStructurePage): void => {
  showCreateDialog({ page })
}
</script>
