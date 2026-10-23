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
    case 'registration':
      return '媒体登记'
    case 'restricted':
      return '受限业务引用'
    default:
      return '其他引用'
  }
}

export const mediaReferenceTypeLabel = (type?: string | null): string => {
  switch (type) {
    case 'restricted_reference':
      return '获授权后可核查'
    case 'product_media':
      return '商品媒体'
    case 'product_category':
      return '商品分类图片'
    case 'product_brand_logo':
      return '商品品牌 Logo'
    case 'product_rich_text':
      return '商品详情与简述'
    case 'product_information_template_rich_text':
      return '商品信息模板正文'
    case 'product_category_rich_text':
      return '商品分类导语'
    case 'product_category_translation_rich_text':
      return '商品分类翻译导语'
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
    case 'workbench_feed_media':
      return '工作台动态图片'
    case 'website_profile_setting':
      return '站点资料'
    case 'site_identity_setting':
      return '站点标识设置'
    case 'site_logo':
      return '站点 Logo'
    case 'site_favicon':
      return '站点 Favicon'
    case 'home_visual_tile':
      return '首页展示图片'
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
    case 'ticket_message_attachment_binding':
      return '工单附件绑定'
    case 'media_upload_session':
      return '附件上传收据'
    case 'media_asset_registration':
      return '原图登记冲突'
    case 'media_derivative_registration':
      return '衍生图登记冲突'
    default:
      return '内容引用'
  }
}
