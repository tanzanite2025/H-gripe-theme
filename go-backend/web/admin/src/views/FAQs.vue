<template>
  <div class="space-y-4">
    <AdminPageHeader title="FAQ 管理" description="维护常见问题、分类、发布状态和展示顺序">
      <template #actions>
        <Button v-if="hasPermission('faq:create')" @click="showCreateDialog">
          <Plus class="size-4" />
          添加 FAQ
        </Button>
      </template>
    </AdminPageHeader>

    <FAQFilterPanel
      :filters="filters"
      :page-filter-options="pageFilterOptions"
      :category-filter-options="categoryFilterOptions"
      :status-filter-options="statusFilterOptions"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <FAQAccordionList
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
      @create-category="openCreateCategoryDialog"
      @edit-category="openEditCategoryDialog"
      @delete-category="requestDeleteCategory"
      @create-faq="openCreateFAQDialog"
    />

    <FAQEditorDialog
      v-model:open="dialogVisible"
      :dialog-mode="dialogMode"
      :faq-form="faqForm"
      :form-errors="formErrors"
      :submitting="submitting"
      :faq-page-options="faqPageOptions"
      :available-faq-categories="availableFAQCategories"
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

    <FAQCategoryEditorDialog
      v-model:open="categoryDialogVisible"
      :mode="categoryDialogMode"
      :category-form="categoryForm"
      :submitting="categorySubmitting"
      :structure-page-options="structurePageOptions"
      :language-options="languageOptions"
      @submit="submitCategoryForm"
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

<script setup>
import { Plus } from '@lucide/vue'
import AdminConfirmDialog from '@/components/admin/AdminConfirmDialog.vue'
import AdminPageHeader from '@/components/admin/AdminPageHeader.vue'
import FAQAccordionList from '@/components/admin/faq/FAQAccordionList.vue'
import FAQCategoryEditorDialog from '@/components/admin/faq/FAQCategoryEditorDialog.vue'
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
  categoryDialogVisible,
  categoryDialogMode,
  categorySubmitting,
  formErrors,
  filters,
  pagination,
  faqForm,
  pageForm,
  categoryForm,
  confirmation,
  statusFilterOptions,
  structureLocales,
  languageOptions,
  structurePageOptions,
  faqPageOptions,
  availableFAQCategories,
  pageFilterOptions,
  categoryFilterOptions,
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
  showCategoryDialog,
  submitCategoryForm,
  isSelected,
  toggleFAQ,
  requestDelete,
  requestBatchDelete,
  requestDeleteCategory,
  executeConfirmedAction
} = useFaqAdmin()

const openCreateCategoryDialog = (page) => {
  showCategoryDialog('create', page)
}

const openEditCategoryDialog = (page, category) => {
  showCategoryDialog('edit', page, category)
}

const openCreateFAQDialog = (page, category) => {
  showCreateDialog({ page, category })
}
</script>
