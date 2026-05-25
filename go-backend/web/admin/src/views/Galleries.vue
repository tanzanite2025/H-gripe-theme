<template>
  <div class="galleries-page">
    <div class="page-header">
      <h2>图库管理</h2>
      <el-button
        v-if="hasPermission('gallery:create')"
        type="primary"
        :icon="Plus"
        @click="showCreateDialog"
      >
        创建图库
      </el-button>
    </div>

    <!-- 图库列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="galleries"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        
        <el-table-column prop="title" label="标题" min-width="200" />
        
        <el-table-column prop="description" label="描述" min-width="250" show-overflow-tooltip />
        
        <el-table-column label="封面" width="120">
          <template #default="{ row }">
            <el-image
              v-if="row.cover_image"
              :src="row.cover_image"
              :preview-src-list="[row.cover_image]"
              fit="cover"
              style="width: 80px; height: 60px; border-radius: 4px;"
            />
          </template>
        </el-table-column>
        
        <el-table-column prop="image_count" label="图片数" width="100" align="center" />
        
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              type="primary"
              size="small"
              link
              @click="viewImages(row)"
            >
              查看图片
            </el-button>
            <el-button
              v-if="hasPermission('gallery:edit')"
              type="primary"
              size="small"
              link
              @click="showEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="hasPermission('gallery:delete')"
              type="danger"
              size="small"
              link
              @click="deleteGallery(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchGalleries"
        @current-change="fetchGalleries"
      />
    </el-card>

    <!-- 创建/编辑图库对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '创建图库' : '编辑图库'"
      width="600px"
    >
      <el-form
        ref="galleryFormRef"
        :model="galleryForm"
        :rules="galleryFormRules"
        label-width="100px"
      >
        <el-form-item label="标题" prop="title">
          <el-input v-model="galleryForm.title" placeholder="请输入标题" />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="galleryForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
          />
        </el-form-item>

        <el-form-item label="封面图片" prop="cover_image">
          <el-input v-model="galleryForm.cover_image" placeholder="请输入封面图片 URL" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">
          确定
        </el-button>
      </template>
    </el-dialog>

    <!-- 图片管理对话框 -->
    <el-dialog
      v-model="imagesDialogVisible"
      :title="`图库: ${currentGallery?.title}`"
      width="900px"
      top="5vh"
    >
      <div class="images-header">
        <el-button
          v-if="hasPermission('gallery:create')"
          type="primary"
          size="small"
          :icon="Plus"
          @click="showAddImageDialog"
        >
          添加图片
        </el-button>
      </div>

      <el-table
        v-loading="imagesLoading"
        :data="images"
        stripe
        style="width: 100%"
        @selection-change="handleImageSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column label="图片" width="120">
          <template #default="{ row }">
            <el-image
              :src="row.url"
              :preview-src-list="images.map(img => img.url)"
              fit="cover"
              style="width: 80px; height: 60px; border-radius: 4px;"
            />
          </template>
        </el-table-column>
        
        <el-table-column prop="title" label="标题" min-width="150" />
        
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        
        <el-table-column prop="sort_order" label="排序" width="80" />
        
        <el-table-column label="操作" width="150" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="hasPermission('gallery:edit')"
              type="primary"
              size="small"
              link
              @click="showEditImageDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="hasPermission('gallery:delete')"
              type="danger"
              size="small"
              link
              @click="deleteImage(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedImages.length > 0" class="batch-actions">
        <span>已选择 {{ selectedImages.length }} 项</span>
        <el-button
          v-if="hasPermission('gallery:delete')"
          type="danger"
          size="small"
          @click="batchDeleteImages"
        >
          批量删除
        </el-button>
      </div>
    </el-dialog>

    <!-- 添加/编辑图片对话框 -->
    <el-dialog
      v-model="imageDialogVisible"
      :title="imageDialogMode === 'create' ? '添加图片' : '编辑图片'"
      width="600px"
    >
      <el-form
        ref="imageFormRef"
        :model="imageForm"
        :rules="imageFormRules"
        label-width="100px"
      >
        <el-form-item label="图片 URL" prop="url">
          <el-input v-model="imageForm.url" placeholder="请输入图片 URL" />
        </el-form-item>

        <el-form-item label="标题" prop="title">
          <el-input v-model="imageForm.title" placeholder="请输入标题" />
        </el-form-item>

        <el-form-item label="描述" prop="description">
          <el-input
            v-model="imageForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入描述"
          />
        </el-form-item>

        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="imageForm.sort_order" :min="0" :max="9999" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="imageDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="imageSubmitting" @click="submitImageForm">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const loading = ref(false)
const galleries = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const galleryFormRef = ref(null)

const imagesDialogVisible = ref(false)
const imagesLoading = ref(false)
const images = ref([])
const currentGallery = ref(null)
const selectedImages = ref([])

const imageDialogVisible = ref(false)
const imageDialogMode = ref('create')
const imageSubmitting = ref(false)
const imageFormRef = ref(null)

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const galleryForm = reactive({
  title: '',
  description: '',
  cover_image: ''
})

const galleryFormRules = {
  title: [
    { required: true, message: '请输入标题', trigger: 'blur' }
  ]
}

const imageForm = reactive({
  url: '',
  title: '',
  description: '',
  sort_order: 0
})

const imageFormRules = {
  url: [
    { required: true, message: '请输入图片 URL', trigger: 'blur' }
  ]
}

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const fetchGalleries = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize
    }

    const response = await axios.get('/api/admin/galleries', { params })
    galleries.value = response.data.galleries
    pagination.total = response.data.total
  } catch (error) {
    ElMessage.error('获取图库列表失败')
  } finally {
    loading.value = false
  }
}

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}

const showEditDialog = async (gallery) => {
  dialogMode.value = 'edit'
  try {
    const response = await axios.get(`/api/admin/galleries/${gallery.id}`)
    Object.assign(galleryForm, response.data.gallery)
    dialogVisible.value = true
  } catch (error) {
    ElMessage.error('获取图库详情失败')
  }
}

const resetForm = () => {
  Object.assign(galleryForm, {
    title: '',
    description: '',
    cover_image: ''
  })
  if (galleryFormRef.value) {
    galleryFormRef.value.clearValidate()
  }
}

const submitForm = async () => {
  if (!galleryFormRef.value) return

  await galleryFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true

    try {
      if (dialogMode.value === 'create') {
        await axios.post('/api/admin/galleries', galleryForm)
        ElMessage.success('图库创建成功')
      } else {
        const { id, ...data } = galleryForm
        await axios.put(`/api/admin/galleries/${id}`, data)
        ElMessage.success('图库更新成功')
      }

      dialogVisible.value = false
      fetchGalleries()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const deleteGallery = async (gallery) => {
  try {
    await ElMessageBox.confirm(`确定要删除图库 "${gallery.title}" 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/galleries/${gallery.id}`)
    ElMessage.success('删除成功')
    fetchGalleries()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const viewImages = async (gallery) => {
  currentGallery.value = gallery
  imagesDialogVisible.value = true
  await fetchImages(gallery.id)
}

const fetchImages = async (galleryId) => {
  imagesLoading.value = true
  try {
    const response = await axios.get(`/api/admin/galleries/${galleryId}/images`)
    images.value = response.data.images
  } catch (error) {
    ElMessage.error('获取图片列表失败')
  } finally {
    imagesLoading.value = false
  }
}

const showAddImageDialog = () => {
  imageDialogMode.value = 'create'
  resetImageForm()
  imageDialogVisible.value = true
}

const showEditImageDialog = (image) => {
  imageDialogMode.value = 'edit'
  Object.assign(imageForm, image)
  imageDialogVisible.value = true
}

const resetImageForm = () => {
  Object.assign(imageForm, {
    url: '',
    title: '',
    description: '',
    sort_order: 0
  })
  if (imageFormRef.value) {
    imageFormRef.value.clearValidate()
  }
}

const submitImageForm = async () => {
  if (!imageFormRef.value) return

  await imageFormRef.value.validate(async (valid) => {
    if (!valid) return

    imageSubmitting.value = true

    try {
      if (imageDialogMode.value === 'create') {
        await axios.post(`/api/admin/galleries/${currentGallery.value.id}/images`, imageForm)
        ElMessage.success('图片添加成功')
      } else {
        const { id, ...data } = imageForm
        await axios.put(`/api/admin/galleries/${currentGallery.value.id}/images/${id}`, data)
        ElMessage.success('图片更新成功')
      }

      imageDialogVisible.value = false
      fetchImages(currentGallery.value.id)
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      imageSubmitting.value = false
    }
  })
}

const deleteImage = async (image) => {
  try {
    await ElMessageBox.confirm(`确定要删除此图片吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/galleries/${currentGallery.value.id}/images/${image.id}`)
    ElMessage.success('删除成功')
    fetchImages(currentGallery.value.id)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleImageSelectionChange = (selection) => {
  selectedImages.value = selection
}

const batchDeleteImages = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedImages.value.length} 张图片吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    const ids = selectedImages.value.map(img => img.id)
    await axios.post(`/api/admin/galleries/${currentGallery.value.id}/images/batch-delete`, { ids })
    ElMessage.success('批量删除成功')
    fetchImages(currentGallery.value.id)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

onMounted(() => {
  fetchGalleries()
})
</script>

<style scoped>
.galleries-page {
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

.table-card {
  margin-bottom: 20px;
}

.images-header {
  margin-bottom: 16px;
}

.batch-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-top: 1px solid #ebeef5;
  margin-top: 12px;
}

.el-pagination {
  margin-top: 20px;
  justify-content: flex-end;
}
</style>
