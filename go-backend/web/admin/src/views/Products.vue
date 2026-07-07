<template>
  <div class="products-page">
    <div class="page-header">
      <h2>商品管理</h2>
      <el-button
        v-if="hasPermission('product:create')"
        type="primary"
        :icon="Plus"
        @click="showCreateDialog"
      >
        添加商品
      </el-button>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">总商品数</div>
            <div class="stat-value">{{ stats.total || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">在售商品</div>
            <div class="stat-value text-success">{{ stats.active || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">低库存</div>
            <div class="stat-value text-warning">{{ stats.low_stock || 0 }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-label">缺货商品</div>
            <div class="stat-value text-danger">{{ stats.out_of_stock || 0 }}</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 筛选栏 -->
    <el-card class="filter-card">
      <el-form :inline="true" :model="filters">
        <el-form-item label="搜索">
          <el-input
            v-model="filters.search"
            placeholder="商品名称/SKU/描述"
            clearable
            @clear="fetchProducts"
            @keyup.enter="fetchProducts"
          />
        </el-form-item>

        <el-form-item label="状态">
          <el-select v-model="filters.status" placeholder="全部" clearable @change="fetchProducts">
            <el-option label="在售" value="active" />
            <el-option label="下架" value="inactive" />
            <el-option label="缺货" value="out_of_stock" />
          </el-select>
        </el-form-item>

        <el-form-item label="语言">
          <el-select v-model="filters.locale" placeholder="全部" clearable @change="fetchProducts">
            <el-option label="中文" value="zh" />
            <el-option label="English" value="en" />
          </el-select>
        </el-form-item>

        <el-form-item label="精选">
          <el-select v-model="filters.featured" placeholder="全部" clearable @change="fetchProducts">
            <el-option label="是" value="true" />
            <el-option label="否" value="false" />
          </el-select>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :icon="Search" @click="fetchProducts">搜索</el-button>
          <el-button :icon="Refresh" @click="resetFilters">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 商品列表 -->
    <el-card class="table-card">
      <el-table
        v-loading="loading"
        :data="products"
        stripe
        style="width: 100%"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        
        <el-table-column prop="id" label="ID" width="80" />
        
        <el-table-column prop="sku" label="SKU" width="120" />
        
        <el-table-column prop="name" label="商品名称" min-width="200" />
        
        <el-table-column prop="price" label="价格" width="120">
          <template #default="{ row }">
            <div>
              <span v-if="row.sale_price" class="sale-price">¥{{ row.sale_price }}</span>
              <span :class="{ 'original-price': row.sale_price }">¥{{ row.price }}</span>
            </div>
          </template>
        </el-table-column>
        
        <el-table-column prop="stock" label="库存" width="100">
          <template #default="{ row }">
            <el-tag v-if="row.stock === 0" type="danger">缺货</el-tag>
            <el-tag v-else-if="row.stock < 10" type="warning">{{ row.stock }}</el-tag>
            <span v-else>{{ row.stock }}</span>
          </template>
        </el-table-column>
        
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusName(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <el-table-column label="精选" width="80">
          <template #default="{ row }">
            <el-icon v-if="row.featured" color="#f59e0b" :size="20">
              <Star />
            </el-icon>
          </template>
        </el-table-column>
        
        <el-table-column prop="locale" label="语言" width="80" />
        
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
        
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button
              v-if="hasPermission('product:edit')"
              type="primary"
              size="small"
              link
              @click="showEditDialog(row)"
            >
              编辑
            </el-button>
            <el-button
              v-if="hasPermission('product:edit')"
              type="warning"
              size="small"
              link
              @click="toggleStatus(row)"
            >
              {{ row.status === 'active' ? '下架' : '上架' }}
            </el-button>
            <el-button
              v-if="hasPermission('product:delete')"
              type="danger"
              size="small"
              link
              @click="deleteProduct(row)"
            >
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 批量操作 -->
      <div v-if="selectedProducts.length > 0" class="batch-actions">
        <span>已选择 {{ selectedProducts.length }} 项</span>
        <div>
          <el-button
            v-if="hasPermission('product:edit')"
            type="success"
            size="small"
            @click="batchUpdateStatus('active')"
          >
            批量上架
          </el-button>
          <el-button
            v-if="hasPermission('product:edit')"
            type="warning"
            size="small"
            @click="batchUpdateStatus('inactive')"
          >
            批量下架
          </el-button>
          <el-button
            v-if="hasPermission('product:delete')"
            type="danger"
            size="small"
            @click="batchDelete"
          >
            批量删除
          </el-button>
        </div>
      </div>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :total="pagination.total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchProducts"
        @current-change="fetchProducts"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogMode === 'create' ? '添加商品' : '编辑商品'"
      width="800px"
    >
      <el-form
        ref="productFormRef"
        :model="productForm"
        :rules="productFormRules"
        label-width="120px"
      >
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="商品名称" prop="name">
              <el-input v-model="productForm.name" placeholder="请输入商品名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="Slug" prop="slug">
              <el-input v-model="productForm.slug" placeholder="请输入 URL slug" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="语言" prop="locale">
              <el-select v-model="productForm.locale" placeholder="请选择语言" style="width: 100%">
                <el-option label="中文" value="zh" />
                <el-option label="English" value="en" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="产品类型">
              <el-select
                v-model="productForm.product_type_id"
                placeholder="请选择产品类型"
                clearable
                style="width: 100%"
                @change="handleProductTypeChange"
              >
                <el-option
                  v-for="type in productTypes"
                  :key="type.id"
                  :label="type.name"
                  :value="type.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <template v-if="selectedSpecDefinitions.length">
          <el-divider content-position="left">规格模板</el-divider>
          <el-row :gutter="20">
            <el-col
              v-for="spec in selectedSpecDefinitions"
              :key="spec.id"
              :span="12"
            >
              <el-form-item :label="getSpecLabel(spec)" :required="spec.is_required">
                <el-input-number
                  v-if="spec.field_type === 'number'"
                  v-model="productForm.specs[spec.slug]"
                  :min="0"
                  style="width: 100%"
                />
                <el-select
                  v-else-if="spec.field_type === 'select'"
                  v-model="productForm.specs[spec.slug]"
                  clearable
                  filterable
                  style="width: 100%"
                >
                  <el-option
                    v-for="option in parseSpecOptions(spec)"
                    :key="option"
                    :label="formatSpecOption(option)"
                    :value="option"
                  />
                </el-select>
                <el-switch
                  v-else-if="spec.field_type === 'boolean'"
                  v-model="productForm.specs[spec.slug]"
                />
                <el-input
                  v-else
                  v-model="productForm.specs[spec.slug]"
                  :placeholder="`请输入${spec.name}`"
                />
              </el-form-item>
            </el-col>
          </el-row>
        </template>

        <el-form-item label="简短描述">
          <el-input
            v-model="productForm.short_description"
            type="textarea"
            :rows="2"
            placeholder="请输入简短描述"
          />
        </el-form-item>

        <el-form-item label="详细描述">
          <el-input
            v-model="productForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入详细描述"
          />
        </el-form-item>

        <el-divider content-position="left">SKU 变体矩阵</el-divider>
        <el-alert
          title="价格、促销价、库存只维护在变体行；商品列表会自动汇总默认变体价格与总库存。"
          type="info"
          show-icon
          :closable="false"
          class="variant-alert"
        />
        <el-table :data="productForm.variants" border size="small" class="variant-table">
          <el-table-column label="默认" width="70" align="center">
            <template #default="{ $index }">
              <el-radio
                :model-value="defaultVariantIndex"
                :label="$index"
                @change="setDefaultVariant($index)"
              >
                &nbsp;
              </el-radio>
            </template>
          </el-table-column>
          <el-table-column label="SKU" min-width="160">
            <template #default="{ row }">
              <el-input v-model="row.sku" placeholder="变体 SKU" />
            </template>
          </el-table-column>
          <el-table-column
            v-for="spec in variantSpecDefinitions"
            :key="spec.id"
            :label="getSpecLabel(spec)"
            min-width="150"
          >
            <template #default="{ row }">
              <el-input-number
                v-if="spec.field_type === 'number'"
                v-model="row.option_values[spec.slug]"
                :min="0"
                style="width: 100%"
              />
              <el-select
                v-else-if="spec.field_type === 'select'"
                v-model="row.option_values[spec.slug]"
                clearable
                filterable
                style="width: 100%"
              >
                <el-option
                  v-for="option in parseSpecOptions(spec)"
                  :key="option"
                  :label="formatSpecOption(option)"
                  :value="option"
                />
              </el-select>
              <el-switch
                v-else-if="spec.field_type === 'boolean'"
                v-model="row.option_values[spec.slug]"
              />
              <el-input
                v-else
                v-model="row.option_values[spec.slug]"
                :placeholder="spec.name"
              />
            </template>
          </el-table-column>
          <el-table-column label="价格" width="130">
            <template #default="{ row }">
              <el-input-number v-model="row.price" :min="0" :precision="2" :step="0.01" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="促销价" width="130">
            <template #default="{ row }">
              <el-input-number v-model="row.sale_price" :min="0" :precision="2" :step="0.01" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="库存" width="120">
            <template #default="{ row }">
              <el-input-number v-model="row.stock" :min="0" style="width: 100%" />
            </template>
          </el-table-column>
          <el-table-column label="启用" width="80" align="center">
            <template #default="{ row }">
              <el-switch v-model="row.is_active" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="80" align="center">
            <template #default="{ $index }">
              <el-button type="danger" link @click="removeVariant($index)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-button class="add-variant-button" type="primary" plain @click="addVariant">
          添加变体
        </el-button>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="重量(克)">
              <el-input-number
                v-model="productForm.weight_grams"
                :min="0"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="状态" prop="status">
              <el-select v-model="productForm.status" placeholder="请选择状态" style="width: 100%">
                <el-option label="在售" value="active" />
                <el-option label="下架" value="inactive" />
                <el-option label="缺货" value="out_of_stock" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="精选商品">
          <el-switch v-model="productForm.featured" />
        </el-form-item>

        <el-form-item label="SEO 标题">
          <el-input v-model="productForm.meta_title" placeholder="请输入 SEO 标题" />
        </el-form-item>

        <el-form-item label="SEO 描述">
          <el-input
            v-model="productForm.meta_description"
            type="textarea"
            :rows="2"
            placeholder="请输入 SEO 描述"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">
          确定
        </el-button>
      </template>
    </el-dialog>

  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Search, Refresh, Star } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import axios from '@/utils/axios'

const authStore = useAuthStore()

const loading = ref(false)
const products = ref([])
const selectedProducts = ref([])
const productTypes = ref([])
const dialogVisible = ref(false)
const dialogMode = ref('create')
const submitting = ref(false)
const productFormRef = ref(null)

const stats = ref({})

const filters = reactive({
  search: '',
  status: '',
  locale: '',
  featured: ''
})

const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

const productForm = reactive({
  id: null,
  product_type_id: null,
  name: '',
  slug: '',
  description: '',
  short_description: '',
  weight_grams: 0,
  status: 'active',
  locale: 'zh',
  featured: false,
  meta_title: '',
  meta_description: '',
  specs: {},
  variants: []
})

const productFormRules = {
  name: [
    { required: true, message: '请输入商品名称', trigger: 'blur' }
  ],
  slug: [
    { required: true, message: '请输入 URL slug', trigger: 'blur' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ]
}

const hasPermission = (permission) => {
  return authStore.hasPermission(permission)
}

const getStatusName = (status) => {
  const statusMap = {
    active: '在售',
    inactive: '下架',
    out_of_stock: '缺货'
  }
  return statusMap[status] || status
}

const getStatusType = (status) => {
  const typeMap = {
    active: 'success',
    inactive: 'info',
    out_of_stock: 'danger'
  }
  return typeMap[status] || ''
}

const formatDate = (dateString) => {
  return new Date(dateString).toLocaleString('zh-CN')
}

const selectedProductType = computed(() => {
  return productTypes.value.find(type => type.id === productForm.product_type_id) || null
})

const selectedSpecDefinitions = computed(() => {
  return (selectedProductType.value?.spec_definitions || [])
    .filter(spec => !spec.is_variant_option)
})

const variantSpecDefinitions = computed(() => {
  return (selectedProductType.value?.spec_definitions || [])
    .filter(spec => spec.is_variant_option)
})

const defaultVariantIndex = computed(() => {
  const index = productForm.variants.findIndex(variant => variant.is_default)
  return index >= 0 ? index : 0
})

const fetchProductTypes = async () => {
  try {
    const response = await axios.get('/api/admin/product-types')
    productTypes.value = response.data?.data || []
  } catch (error) {
    console.error('获取产品类型失败', error)
  }
}

const parseSpecOptions = (spec) => {
  if (!spec?.options) return []
  try {
    const parsed = JSON.parse(spec.options)
    return Array.isArray(parsed) ? parsed : []
  } catch (error) {
    return []
  }
}

const formatSpecOption = (option) => {
  return String(option).replace(/_/g, ' ')
}

const getSpecLabel = (spec) => {
  return spec.unit ? `${spec.name} (${spec.unit})` : spec.name
}

const coerceSpecValueForForm = (definition, value) => {
  if (!definition) return value
  if (definition.field_type === 'number') {
    const numberValue = Number(value)
    return Number.isFinite(numberValue) ? numberValue : undefined
  }
  if (definition.field_type === 'boolean') {
    return value === true || value === 'true' || value === '1'
  }
  return value
}

const buildSpecFormValues = (product) => {
  const values = {}
  ;(product.spec_values || []).forEach((item) => {
    const definition = item.definition
    if (!definition?.slug) return
    values[definition.slug] = coerceSpecValueForForm(definition, item.value)
  })
  return values
}

const parseVariantOptions = (variant) => {
  if (!variant?.option_values) return {}
  if (typeof variant.option_values === 'object') return { ...variant.option_values }
  try {
    const parsed = JSON.parse(variant.option_values)
    return parsed && typeof parsed === 'object' ? parsed : {}
  } catch (error) {
    return {}
  }
}

const createEmptyVariant = (overrides = {}) => ({
  id: null,
  sku: '',
  title: '',
  option_values: {},
  price: 0,
  sale_price: null,
  stock: 0,
  weight_grams: 0,
  is_default: false,
  is_active: true,
  sort_order: productForm.variants.length * 10,
  ...overrides
})

const buildVariantFormValues = (product) => {
  const variants = (product.variants || []).map((variant, index) => createEmptyVariant({
    id: variant.id || null,
    sku: variant.sku || '',
    title: variant.title || '',
    option_values: parseVariantOptions(variant),
    price: Number(variant.price || 0),
    sale_price: variant.sale_price ?? null,
    stock: Number(variant.stock || 0),
    weight_grams: variant.weight_grams ?? variant.weight ?? 0,
    is_default: Boolean(variant.is_default),
    is_active: variant.is_active !== false,
    sort_order: variant.sort_order ?? index * 10
  }))

  if (variants.length === 0) {
    variants.push(createEmptyVariant({
      weight_grams: product.weight_grams ?? product.weight ?? 0,
      is_default: true
    }))
  }

  if (!variants.some(variant => variant.is_default)) {
    variants[0].is_default = true
  }
  return variants
}

const addVariant = () => {
  productForm.variants.push(createEmptyVariant({
    is_default: productForm.variants.length === 0
  }))
}

const removeVariant = (index) => {
  if (productForm.variants.length <= 1) {
    ElMessage.warning('至少保留一个变体')
    return
  }
  const wasDefault = productForm.variants[index]?.is_default
  productForm.variants.splice(index, 1)
  if (wasDefault && productForm.variants.length > 0) {
    setDefaultVariant(0)
  }
}

const setDefaultVariant = (index) => {
  productForm.variants.forEach((variant, currentIndex) => {
    variant.is_default = currentIndex === index
  })
}

const normalizeFormVariants = () => {
  const variants = productForm.variants
  if (!variants.length) {
    return []
  }

  if (!variants.some(variant => variant.is_default)) {
    variants[0].is_default = true
  }

  return variants.map((variant, index) => {
    const optionValues = {}
    variantSpecDefinitions.value.forEach((spec) => {
      const value = variant.option_values?.[spec.slug]
      if (value !== undefined && value !== null && value !== '') {
        optionValues[spec.slug] = value
      }
    })

    return {
      id: variant.id || undefined,
      sku: String(variant.sku || '').trim(),
      title: String(variant.title || '').trim(),
      option_values: optionValues,
      price: Number(variant.price || 0),
      sale_price: variant.sale_price === '' ? null : variant.sale_price,
      stock: Number(variant.stock || 0),
      weight_grams: Number(variant.weight_grams || productForm.weight_grams || 0),
      is_default: Boolean(variant.is_default),
      is_active: variant.is_active !== false,
      sort_order: Number(variant.sort_order ?? index * 10)
    }
  })
}

const buildProductPayload = () => {
  const variants = normalizeFormVariants()
  const {
    id,
    product_type_id,
    name,
    slug,
    description,
    short_description,
    weight_grams,
    status,
    locale,
    parent_id,
    featured,
    meta_title,
    meta_description,
    specs
  } = productForm

  return {
    id,
    product_type_id,
    name,
    slug,
    description,
    short_description,
    weight_grams,
    status,
    locale,
    parent_id,
    featured,
    meta_title,
    meta_description,
    specs,
    variants
  }
}

const validateVariants = (variants) => {
  if (!variants.length) {
    ElMessage.error('请至少添加一个 SKU 变体')
    return false
  }
  const invalidSKU = variants.find(variant => !variant.sku)
  if (invalidSKU) {
    ElMessage.error('每个变体都必须填写 SKU')
    return false
  }
  const invalidPrice = variants.find(variant => Number(variant.price || 0) <= 0)
  if (invalidPrice) {
    ElMessage.error('每个变体价格必须大于 0')
    return false
  }
  const invalidStock = variants.find(variant => Number(variant.stock || 0) < 0)
  if (invalidStock) {
    ElMessage.error('变体库存不能为负数')
    return false
  }
  return true
}

const handleProductTypeChange = () => {
  const nextSpecs = {}
  selectedSpecDefinitions.value.forEach((spec) => {
    if (spec.field_type === 'boolean') {
      nextSpecs[spec.slug] = false
    }
  })
  productForm.specs = nextSpecs
  productForm.variants.forEach((variant) => {
    variant.option_values = {}
  })
}

const fetchStats = async () => {
  try {
    const response = await axios.get('/api/admin/products/stats')
    stats.value = response.data
  } catch (error) {
    console.error('获取统计数据失败', error)
  }
}

const fetchProducts = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      page_size: pagination.pageSize,
      ...filters
    }

    const response = await axios.get('/api/admin/products', { params })
    products.value = response.data.products
    pagination.total = response.data.pagination.total
  } catch (error) {
    ElMessage.error('获取商品列表失败')
  } finally {
    loading.value = false
  }
}

const resetFilters = () => {
  filters.search = ''
  filters.status = ''
  filters.locale = ''
  filters.featured = ''
  pagination.page = 1
  fetchProducts()
}

const showCreateDialog = () => {
  dialogMode.value = 'create'
  resetForm()
  dialogVisible.value = true
}

const showEditDialog = async (product) => {
  dialogMode.value = 'edit'
  let detail = product
  try {
    if (productTypes.value.length === 0) {
      await fetchProductTypes()
    }
    const response = await axios.get(`/api/admin/products/${product.id}`)
    detail = response.data?.product || product
  } catch (error) {
    ElMessage.warning('获取商品详情失败，使用列表数据编辑')
  }

  Object.assign(productForm, {
    id: detail.id,
    product_type_id: detail.product_type_id || detail.product_type?.id || null,
    name: detail.name,
    slug: detail.slug,
    description: detail.description || '',
    short_description: detail.short_description || detail.short_desc || '',
    weight_grams: detail.weight_grams ?? detail.weight ?? 0,
    status: detail.status,
    locale: detail.locale || 'zh',
    featured: detail.featured || false,
    meta_title: detail.meta_title || '',
    meta_description: detail.meta_description || detail.meta_desc || '',
    specs: buildSpecFormValues(detail),
    variants: buildVariantFormValues(detail)
  })
  dialogVisible.value = true
}

const resetForm = () => {
  Object.assign(productForm, {
    id: null,
    product_type_id: null,
    name: '',
    slug: '',
    description: '',
    short_description: '',
    weight_grams: 0,
    status: 'active',
    locale: 'zh',
    featured: false,
    meta_title: '',
    meta_description: '',
    specs: {},
    variants: [createEmptyVariant({ is_default: true })]
  })
  if (productFormRef.value) {
    productFormRef.value.clearValidate()
  }
}

const submitForm = async () => {
  if (!productFormRef.value) return

  await productFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true

    try {
      const payload = buildProductPayload()
      if (!validateVariants(payload.variants)) {
        return
      }
      if (dialogMode.value === 'create') {
        await axios.post('/api/admin/products', payload)
        ElMessage.success('商品创建成功')
      } else {
        const { id, ...data } = payload
        await axios.put(`/api/admin/products/${id}`, data)
        ElMessage.success('商品更新成功')
      }

      dialogVisible.value = false
      fetchProducts()
      fetchStats()
    } catch (error) {
      ElMessage.error(error.response?.data?.error || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

const toggleStatus = async (product) => {
  const newStatus = product.status === 'active' ? 'inactive' : 'active'
  const action = newStatus === 'active' ? '上架' : '下架'

  try {
    await ElMessageBox.confirm(`确定要${action}商品 ${product.name} 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    await axios.patch(`/api/admin/products/${product.id}/status`, { status: newStatus })
    ElMessage.success(`${action}成功`)
    fetchProducts()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}失败`)
    }
  }
}

const deleteProduct = async (product) => {
  try {
    await ElMessageBox.confirm(`确定要删除商品 ${product.name} 吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    await axios.delete(`/api/admin/products/${product.id}`)
    ElMessage.success('删除成功')
    fetchProducts()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

const handleSelectionChange = (selection) => {
  selectedProducts.value = selection
}

const batchUpdateStatus = async (status) => {
  const action = status === 'active' ? '上架' : '下架'
  
  try {
    await ElMessageBox.confirm(`确定要${action}选中的 ${selectedProducts.value.length} 个商品吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })

    const productIds = selectedProducts.value.map(p => p.id)
    await axios.post('/api/admin/products/batch-status', {
      product_ids: productIds,
      status: status
    })
    ElMessage.success(`批量${action}成功`)
    fetchProducts()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`批量${action}失败`)
    }
  }
}

const batchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedProducts.value.length} 个商品吗？此操作不可恢复！`, '警告', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'error'
    })

    const productIds = selectedProducts.value.map(p => p.id)
    await axios.post('/api/admin/products/batch-delete', { product_ids: productIds })
    ElMessage.success('批量删除成功')
    fetchProducts()
    fetchStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

onMounted(() => {
  fetchProductTypes()
  fetchStats()
  fetchProducts()
})
</script>

<style scoped>
.products-page {
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

.stats-row {
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
}

.stat-label {
  font-size: 14px;
  color: #909399;
  margin-bottom: 8px;
}

.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
}

.text-success {
  color: #67c23a;
}

.text-warning {
  color: #e6a23c;
}

.text-danger {
  color: #f56c6c;
}

.filter-card {
  margin-bottom: 20px;
}

.table-card {
  margin-bottom: 20px;
}

.sale-price {
  color: #f56c6c;
  font-weight: bold;
  margin-right: 8px;
}

.original-price {
  text-decoration: line-through;
  color: #909399;
  font-size: 12px;
}

.variant-alert {
  margin-bottom: 12px;
}

.variant-table {
  margin-bottom: 12px;
}

.add-variant-button {
  margin-bottom: 20px;
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
