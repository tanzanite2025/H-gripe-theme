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

    <FAQStructurePanel
      :structure-locales="structureLocales"
      :active-structure-locale="activeStructureLocale"
      :loading="structureLoading"
      :faq-structure="faqStructure"
      :has-permission="hasPermission"
      :domain-name="domainName"
      :visibility-name="visibilityName"
      :visibility-tone="visibilityTone"
      @switch-locale="switchStructureLocale"
      @edit-page="showPageDialog"
      @create-category="openCreateCategoryDialog"
      @edit-category="openEditCategoryDialog"
      @delete-category="requestDeleteCategory"
    />

    <FAQFilterPanel
      :filters="filters"
      :page-filter-options="pageFilterOptions"
      :category-filter-options="categoryFilterOptions"
      :status-filter-options="statusFilterOptions"
      :locale-filter-options="localeFilterOptions"
      @apply="applyFilters"
      @reset="resetFilters"
    />

    <FAQTable
      :loading="loading"
      :faqs="faqs"
      :selected-faqs="selectedFAQs"
      :selection-state="selectionState"
      :pagination="pagination"
      :has-permission="hasPermission"
      :is-selected="isSelected"
      :plain-text="plainTextFromHTML"
      :page-title="pageTitleForFAQ"
      :category-label="categoryLabelForFAQ"
      :status-tone="statusTone"
      :status-name="statusName"
      :locale-name="localeName"
      :format-date="formatDate"
      @toggle-all="toggleAllFAQs"
      @toggle-faq="toggleFAQ"
      @edit="showEditDialog"
      @delete="requestDelete"
      @batch-delete="requestBatchDelete"
      @update-page="updatePage"
      @update-page-size="updatePageSize"
    />

    <FAQEditorDialog
      v-model:open="dialogVisible"
      :dialog-mode="dialogMode"
      :faq-form="faqForm"
      :form-errors="formErrors"
      :submitting="submitting"
      :faq-page-options="faqPageOptions"
      :available-faq-categories="availableFAQCategories"
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
import FAQCategoryEditorDialog from '@/components/admin/faq/FAQCategoryEditorDialog.vue'
import FAQEditorDialog from '@/components/admin/faq/FAQEditorDialog.vue'
import FAQFilterPanel from '@/components/admin/faq/FAQFilterPanel.vue'
import FAQPageEditorDialog from '@/components/admin/faq/FAQPageEditorDialog.vue'
import FAQStructurePanel from '@/components/admin/faq/FAQStructurePanel.vue'
import FAQTable from '@/components/admin/faq/FAQTable.vue'
import { Button } from '@/components/ui/button'
import { useFaqAdmin } from '@/composables/useFaqAdmin'

const {
  loading,
  structureLoading,
  faqs,
  activeStructureLocale,
  selectedFAQs,
  dialogVisible,
  dialogMode,
  submitting,
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
  localeFilterOptions,
  structureLocales,
  faqStructure,
  structurePageOptions,
  faqPageOptions,
  availableFAQCategories,
  pageFilterOptions,
  categoryFilterOptions,
  selectionState,
  hasPermission,
  localeName,
  statusName,
  statusTone,
  visibilityName,
  visibilityTone,
  domainName,
  formatDate,
  plainTextFromHTML,
  pageTitleForFAQ,
  categoryLabelForFAQ,
  clearFieldError,
  updateFAQAnswer,
  switchStructureLocale,
  applyFilters,
  resetFilters,
  updatePage,
  updatePageSize,
  showCreateDialog,
  showEditDialog,
  submitForm,
  showPageDialog,
  submitPageForm,
  showCategoryDialog,
  submitCategoryForm,
  isSelected,
  toggleAllFAQs,
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
</script>
