package service

import (
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/outbox"
)

type DashboardBusinessFollowUpTaskDefinition struct {
	Title           string
	Category        string
	Impact          string
	ResponsibleTeam string
	SalesNextSteps  []string
	ReadPermission  auth.Permission
}

func buildDashboardBusinessFollowUpEmailTaskDefinition(title, impact string) DashboardBusinessFollowUpTaskDefinition {
	return DashboardBusinessFollowUpTaskDefinition{Title: title, Category: "客户通知", Impact: impact, ResponsibleTeam: "客户服务 / 邮件服务负责人", ReadPermission: auth.PermOrderView,
		SalesNextSteps: []string{"先打开订单，核实付款、发货或售后记录的实际状态。", "客户正在等待时，可通过已确认的联系方式说明实际进展，并记录已联系的情况。", "自动尝试已停止或客户急需通知时，将此跟进说明交给邮件服务负责人；补发前确认是否已经另行通知客户。"}}
}

var dashboardBusinessFollowUpTaskDefinitions = map[string]DashboardBusinessFollowUpTaskDefinition{
	outbox.EventTypeOrderPaymentSucceeded:           buildDashboardBusinessFollowUpEmailTaskDefinition("付款确认邮件发送未完成", "客户可能尚未收到付款确认。此事项反映通知任务出错，付款是否成功请以订单和支付记录为准，勿据此重复收款。"),
	outbox.EventTypeOrderPaymentExpired:             buildDashboardBusinessFollowUpEmailTaskDefinition("付款到期提醒邮件发送未完成", "客户可能不知道原付款已经到期；沟通前请确认订单当前是否仍可付款。"),
	outbox.EventTypeOrderCancelled:                  buildDashboardBusinessFollowUpEmailTaskDefinition("订单取消通知邮件发送未完成", "客户可能尚未收到取消通知；订单是否已取消、是否涉及退款，请核实订单记录。"),
	outbox.EventTypeOrderShipped:                    buildDashboardBusinessFollowUpEmailTaskDefinition("发货通知邮件发送未完成", "客户可能尚未收到发货通知和查询物流的入口。邮件任务失败不能据此判断包裹未发出，请查看实际发货记录。"),
	outbox.EventTypeOrderDelivered:                  buildDashboardBusinessFollowUpEmailTaskDefinition("签收通知邮件发送未完成", "客户可能尚未收到签收通知；包裹是否妥投请核对物流记录。"),
	outbox.EventTypeOrderRefunded:                   buildDashboardBusinessFollowUpEmailTaskDefinition("退款通知邮件发送未完成", "客户可能尚未收到退款通知。邮件未完成不能据此判断退款失败或到账，请核对退款和支付渠道记录。"),
	outbox.EventTypeAfterSalesStatusChanged:         buildDashboardBusinessFollowUpEmailTaskDefinition("售后进展通知邮件发送未完成", "客户可能不知道退换货案件的最新进展；需要核实当前审核、寄回或处理结果并及时沟通。"),
	outbox.EventTypeOrderDisputeContactEmail:        buildDashboardBusinessFollowUpEmailTaskDefinition("拒付沟通邮件发送未完成", "针对拒付争议的沟通邮件尚未完成。争议的应诉期限不会因此延长，需要在争议处理页核对期限和处理进展。"),
	outbox.EventTypeOrderPaid:                       {Title: "已付款订单向外部系统同步未完成", Category: "订单同步", Impact: "已付款订单的通知没有完成向配置的外部系统同步，下游系统可能尚未收到该订单。", ResponsibleTeam: "订单对接负责人", ReadPermission: auth.PermOrderView, SalesNextSteps: []string{"打开订单，核实付款和当前履约进度。", "联系订单对接负责人核实下游是否已收到订单，避免重复建立订单或重复发货。"}},
	outbox.EventTypeOrderCompleted:                  {Title: "订单完成后的积分或通知任务未完成", Category: "积分与通知", Impact: "订单完成后的积分登记或客户通知可能尚未完成；不能仅凭这项任务判断是哪一环失败，需分别核对积分和通知记录。", ResponsibleTeam: "会员积分 / 客户服务负责人", ReadPermission: auth.PermOrderView, SalesNextSteps: []string{"打开订单并核对积分及客户通知记录。", "将未完成的具体事项交给对应负责人处理，向客户说明已核实的进展。"}},
	outbox.EventTypeTrackingShipmentRegistration:    {Title: "包裹物流查询登记未完成", Category: "物流跟进", Impact: "包裹可能尚未成功登记到物流查询服务，客户可能暂时看不到轨迹；这不代表包裹没有发出或已经丢失。", ResponsibleTeam: "物流负责人", ReadPermission: auth.PermOrderView, SalesNextSteps: []string{"打开订单核对承运商、运单号和实际交运情况。", "向物流负责人提供运单信息，核查承运商匹配和查询服务配置。", "恢复后核实客户能否查询轨迹，并告知客户真实物流进展。"}},
	outbox.EventTypePaymentRefundExecutionRequested: {Title: "退款执行任务需要核查", Category: "退款跟进", Impact: "后台退款执行任务出错，无法据此确认退款是否到账；需要结合退款记录和支付渠道结果判断。", ResponsibleTeam: "财务 / 支付负责人", ReadPermission: auth.PermOrderView, SalesNextSteps: []string{"打开对应订单，核对退款申请及退款执行记录。", "将说明交给财务或支付负责人，先核查渠道是否已经退款，再决定如何处理。", "向客户说明已确认的处理进度，渠道结果确认后再承诺到账情况。"}},
	outbox.EventTypeReferralOrderPaid:               {Title: "付款后的推荐奖励计算未完成", Category: "会员奖励", Impact: "与这笔订单相关的推荐奖励计算可能延迟，奖励是否已发放需要查会员台账。", ResponsibleTeam: "会员奖励负责人", ReadPermission: auth.PermOrderView, SalesNextSteps: []string{"核对订单付款状态及推荐奖励台账。", "向会员奖励负责人提供订单信息，补处理前确认是否已经发奖。"}},
	outbox.EventTypeReferralOrderDelivered:          {Title: "签收后的推荐奖励处理未完成", Category: "会员奖励", Impact: "订单签收后的推荐奖励确认可能延迟。", ResponsibleTeam: "会员奖励负责人", ReadPermission: auth.PermOrderView, SalesNextSteps: []string{"核对订单签收及会员奖励记录。", "联系会员奖励负责人确认待处理奖励，避免重复发放。"}},
	outbox.EventTypeReferralOrderInvalidated:        {Title: "订单变更后的推荐奖励调整未完成", Category: "会员奖励", Impact: "订单取消或退款后的推荐奖励撤销、调整可能尚未完成。", ResponsibleTeam: "会员奖励负责人", ReadPermission: auth.PermOrderView, SalesNextSteps: []string{"核对订单变更与会员奖励台账。", "联系会员奖励负责人完成奖励调整，不直接依据本清单修改客户积分。"}},
	outbox.EventTypeVerifiedConversion:              {Title: "已付款订单的营销转化同步未完成", Category: "营销同步", Impact: "订单付款后的营销转化数据同步可能延迟；订单是否已付款仍以订单记录为准。", ResponsibleTeam: "营销对接负责人", ReadPermission: auth.PermOrderView, SalesNextSteps: []string{"核对订单付款记录。", "将跟进说明交给营销对接负责人核查外部接收结果。"}},
	outbox.EventTypeMerchantProductUpsert:           {Title: "商品向 Google Merchant 更新未完成", Category: "商品渠道", Impact: "渠道展示的商品信息可能尚未更新，需核对实际渠道发布结果。", ResponsibleTeam: "商品渠道负责人", ReadPermission: auth.PermMerchantView, SalesNextSteps: []string{"核对商品及 Google Merchant 发布记录。", "联系渠道负责人检查商品资料和渠道授权，恢复后确认渠道展示结果。"}},
	outbox.EventTypeMerchantProductWithdraw:         {Title: "商品在 Google Merchant 下架未完成", Category: "商品渠道", Impact: "渠道中可能仍展示应下架的商品，需要核实实际可售状态。", ResponsibleTeam: "商品渠道负责人", ReadPermission: auth.PermMerchantView, SalesNextSteps: []string{"核对商品可售状态和渠道展示。", "联系渠道负责人优先确认下架结果，避免继续推广不可售商品。"}},
	outbox.EventTypeMerchantOfferRevalidate:         {Title: "渠道商品资料复核未完成", Category: "商品渠道", Impact: "渠道商品的发布资料复核尚未完成，不能据此确认已通过渠道审核。", ResponsibleTeam: "商品渠道负责人", ReadPermission: auth.PermMerchantView, SalesNextSteps: []string{"打开渠道发布记录，核对资料和校验结果。", "联系渠道负责人补齐资料并重新核实发布结果。"}},
	outbox.EventTypeProductCacheInvalidate:          {Title: "商品页面更新任务未完成", Category: "商品页面", Impact: "商品更新后的页面刷新任务出错，客户可能仍看到旧资料；需要核对前台实际内容。", ResponsibleTeam: "商品 / 系统负责人", ReadPermission: auth.PermProductView, SalesNextSteps: []string{"核对后台商品资料与客户实际看到的页面是否一致。", "存在差异时将具体商品链接交给商品或系统负责人处理，恢复后再次核查页面。"}},
	outbox.EventTypeCustomerServiceRealtime:         {Title: "客服会话实时更新未完成", Category: "客服沟通", Impact: "客服会话的实时更新任务出错，界面可能未及时显示新消息；历史消息记录仍需在会话中核对。", ResponsibleTeam: "客服 / 系统负责人", ReadPermission: auth.PermTicketView, SalesNextSteps: []string{"进入客服会话并刷新，核查客户最新消息和已发送回复。", "客户正在等待时先通过会话继续跟进，持续不能更新时联系系统负责人。"}},
}
