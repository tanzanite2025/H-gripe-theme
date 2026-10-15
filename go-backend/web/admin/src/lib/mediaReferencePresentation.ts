export const mediaReferenceCategoryLabel = (category?: string | null): string => {
  switch (category) {
    case 'catalog':
      return '商品与目录'
    case 'content':
      return '内容页面'
    case 'community':
      return '社区内容'
    case 'customer':
      return '客户记录'
    case 'support':
      return '客服与工单'
    default:
      return '其他引用'
  }
}

export const mediaReferenceTypeLabel = (type?: string | null): string => {
  switch (type) {
    case 'product_media':
      return '商品媒体'
    case 'product_category':
      return '商品分类图片'
    case 'product_rich_text':
      return '商品详情与简述'
    case 'product_option_swatch':
      return '商品选项色板'
    case 'product_template_option_swatch':
      return '规格模板色板'
    case 'faq':
      return 'FAQ'
    case 'gallery_cover':
      return '图库封面'
    case 'gallery_image':
      return '图库图片'
    case 'post':
      return '文章内容'
    case 'showcase':
      return '买家秀'
    case 'review':
      return '商品评价'
    case 'warranty_claim':
      return '保修申请'
    case 'suggestion_feedback':
      return '建议反馈'
    case 'ticket_auto_reply':
      return '客服自动回复'
    case 'ticket_message':
      return '工单消息'
    case 'media_upload_session':
      return '待发送附件'
    default:
      return '内容引用'
  }
}
