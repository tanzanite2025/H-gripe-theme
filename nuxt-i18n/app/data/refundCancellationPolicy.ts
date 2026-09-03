import { normalizeStorefrontLocaleCode } from '~/utils/storefrontLocales'

export interface RefundCancellationPolicyWindow {
  id: string
  label: string
  title: string
  body: string
}

export interface RefundCancellationPolicyContent {
  title: string
  intro: string
  windows: RefundCancellationPolicyWindow[]
}

const refundCancellationPolicyByLocale: Record<string, RefundCancellationPolicyContent> = {
  en: {
    title: 'Refund & Cancellation Policy',
    intro: 'Cancellation requests for custom orders are handled by the payment and fulfillment stage.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Within 24 hours after payment',
        title: 'Free cancellation',
        body: 'Cancel within 24 hours after payment for a 100% full refund to your original payment method.',
      },
      {
        id: 'after-24-hours',
        label: 'After 24 hours after payment (scheduled for production/cutting)',
        title: 'Custom handling fee applies',
        body: 'Once the order has entered production scheduling or material cutting, a 15%-20% Custom Handling / Restocking Fee will be deducted for custom materials and labor already committed.',
      },
      {
        id: 'shipped',
        label: 'Shipped',
        title: 'Use the standard return process',
        body: 'Direct cancellation is no longer accepted after shipment. Please receive the package first, then follow the regular return process if eligible.',
      },
    ],
  },
  zh_cn: {
    title: '退款与取消政策',
    intro: '定制订单的取消请求将按付款时间和履约阶段处理。',
    windows: [
      {
        id: 'within-24-hours',
        label: '付款后 24 小时内',
        title: '免费取消',
        body: '付款后 24 小时内取消，可获得 100% 全额原路退款。',
      },
      {
        id: 'after-24-hours',
        label: '付款 24 小时后（已进入排产/下料）',
        title: '将扣除定制处理费用',
        body: '订单进入排产或下料后，将扣除 15%~20% 的 Custom Handling / Restocking Fee，用于已投入的定制材料与工时。',
      },
      {
        id: 'shipped',
        label: '已发货',
        title: '按常规退货流程处理',
        body: '发货后不接受直接取消。请先完成收货，符合条件的订单需按收货后的常规退货流程处理。',
      },
    ],
  },
  fr: {
    title: 'Politique de remboursement et d’annulation',
    intro: 'Les demandes d’annulation pour les commandes personnalisées sont traitées selon le moment du paiement et l’étape d’exécution.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Dans les 24 heures suivant le paiement',
        title: 'Annulation gratuite',
        body: 'Une annulation dans les 24 heures suivant le paiement donne droit à un remboursement intégral de 100 % sur le mode de paiement initial.',
      },
      {
        id: 'after-24-hours',
        label: 'Plus de 24 heures après le paiement (production/découpe planifiée)',
        title: 'Frais de traitement personnalisé applicables',
        body: 'Une fois la commande entrée en planification de production ou en découpe de matériaux, des frais Custom Handling / Restocking Fee de 15 % à 20 % seront déduits pour les matériaux personnalisés et le temps de travail déjà engagés.',
      },
      {
        id: 'shipped',
        label: 'Expédiée',
        title: 'Utiliser la procédure de retour standard',
        body: 'L’annulation directe n’est plus acceptée après l’expédition. Veuillez d’abord recevoir le colis, puis suivre la procédure de retour habituelle si la commande est éligible.',
      },
    ],
  },
  de: {
    title: 'Rückerstattungs- und Stornierungsrichtlinie',
    intro: 'Stornierungsanfragen für kundenspezifische Bestellungen werden nach Zahlungszeitpunkt und Bearbeitungsstatus behandelt.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Innerhalb von 24 Stunden nach Zahlung',
        title: 'Kostenlose Stornierung',
        body: 'Bei Stornierung innerhalb von 24 Stunden nach Zahlung erhalten Sie eine vollständige Rückerstattung von 100 % auf die ursprüngliche Zahlungsmethode.',
      },
      {
        id: 'after-24-hours',
        label: 'Nach 24 Stunden nach Zahlung (für Produktion/Zuschnitt eingeplant)',
        title: 'Gebühr für kundenspezifische Bearbeitung',
        body: 'Sobald die Bestellung in die Produktionsplanung oder den Materialzuschnitt aufgenommen wurde, wird eine Custom Handling / Restocking Fee von 15 %-20 % für bereits gebundene kundenspezifische Materialien und Arbeitszeit abgezogen.',
      },
      {
        id: 'shipped',
        label: 'Versendet',
        title: 'Standard-Rückgabeprozess nutzen',
        body: 'Nach dem Versand wird keine direkte Stornierung mehr akzeptiert. Bitte nehmen Sie das Paket zuerst an und folgen Sie danach, sofern berechtigt, dem regulären Rückgabeprozess.',
      },
    ],
  },
  es: {
    title: 'Política de reembolso y cancelación',
    intro: 'Las solicitudes de cancelación de pedidos personalizados se gestionan según el momento del pago y la etapa de preparación.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Dentro de las 24 horas posteriores al pago',
        title: 'Cancelación gratuita',
        body: 'Si cancelas dentro de las 24 horas posteriores al pago, recibirás un reembolso completo del 100 % al método de pago original.',
      },
      {
        id: 'after-24-hours',
        label: 'Después de 24 horas del pago (programado para producción/corte)',
        title: 'Se aplica una tarifa de manejo personalizado',
        body: 'Una vez que el pedido haya entrado en programación de producción o corte de materiales, se deducirá una Custom Handling / Restocking Fee del 15 %-20 % por materiales personalizados y mano de obra ya comprometidos.',
      },
      {
        id: 'shipped',
        label: 'Enviado',
        title: 'Usar el proceso de devolución estándar',
        body: 'Después del envío no se acepta la cancelación directa. Recibe primero el paquete y luego sigue el proceso regular de devolución si el pedido cumple los requisitos.',
      },
    ],
  },
  ja: {
    title: '返金およびキャンセルポリシー',
    intro: 'カスタム注文のキャンセルリクエストは、お支払い時点と履行段階に応じて処理されます。',
    windows: [
      {
        id: 'within-24-hours',
        label: 'お支払い後24時間以内',
        title: '無料キャンセル',
        body: 'お支払い後24時間以内のキャンセルは、元のお支払い方法へ100%全額返金されます。',
      },
      {
        id: 'after-24-hours',
        label: 'お支払いから24時間経過後（生産手配/材料カット開始）',
        title: 'カスタム処理手数料が適用されます',
        body: '注文が生産手配または材料カットに入った後は、すでに確保したカスタム材料と工数に対して15%~20%の Custom Handling / Restocking Fee を差し引きます。',
      },
      {
        id: 'shipped',
        label: '発送済み',
        title: '通常の返品手続きに従ってください',
        body: '発送後は直接キャンセルを受け付けていません。まず商品をお受け取りいただき、対象となる場合は受け取り後の通常の返品手続きに従ってください。',
      },
    ],
  },
  ko: {
    title: '환불 및 취소 정책',
    intro: '맞춤 주문의 취소 요청은 결제 시점과 이행 단계에 따라 처리됩니다.',
    windows: [
      {
        id: 'within-24-hours',
        label: '결제 후 24시간 이내',
        title: '무료 취소',
        body: '결제 후 24시간 이내에 취소하면 원래 결제 수단으로 100% 전액 환불됩니다.',
      },
      {
        id: 'after-24-hours',
        label: '결제 후 24시간 경과 (생산 일정/자재 절단 단계)',
        title: '맞춤 처리 수수료 적용',
        body: '주문이 생산 일정 또는 자재 절단 단계에 들어가면 이미 투입된 맞춤 자재와 작업 시간에 대해 15%~20%의 Custom Handling / Restocking Fee가 차감됩니다.',
      },
      {
        id: 'shipped',
        label: '배송 완료',
        title: '일반 반품 절차 이용',
        body: '배송 후에는 직접 취소가 허용되지 않습니다. 먼저 상품을 수령한 뒤, 조건에 해당하는 경우 수령 후 일반 반품 절차에 따라 진행해 주세요.',
      },
    ],
  },
  it: {
    title: 'Politica di rimborso e cancellazione',
    intro: 'Le richieste di cancellazione per ordini personalizzati sono gestite in base al momento del pagamento e alla fase di evasione.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Entro 24 ore dal pagamento',
        title: 'Cancellazione gratuita',
        body: 'La cancellazione entro 24 ore dal pagamento dà diritto a un rimborso completo del 100 % sul metodo di pagamento originale.',
      },
      {
        id: 'after-24-hours',
        label: 'Dopo 24 ore dal pagamento (programmazione produzione/taglio)',
        title: 'Si applica una commissione di gestione personalizzata',
        body: 'Una volta che l’ordine è entrato nella programmazione della produzione o nel taglio dei materiali, verrà detratta una Custom Handling / Restocking Fee del 15 %-20 % per materiali personalizzati e manodopera già impegnati.',
      },
      {
        id: 'shipped',
        label: 'Spedito',
        title: 'Usare la procedura di reso standard',
        body: 'Dopo la spedizione non è più accettata la cancellazione diretta. Ricevi prima il pacco, poi segui la normale procedura di reso se l’ordine è idoneo.',
      },
    ],
  },
  pt: {
    title: 'Política de reembolso e cancelamento',
    intro: 'Os pedidos de cancelamento de encomendas personalizadas são tratados de acordo com o momento do pagamento e a fase de processamento.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'No prazo de 24 horas após o pagamento',
        title: 'Cancelamento gratuito',
        body: 'Se cancelar no prazo de 24 horas após o pagamento, receberá um reembolso integral de 100 % para o método de pagamento original.',
      },
      {
        id: 'after-24-hours',
        label: 'Após 24 horas do pagamento (produção/corte programados)',
        title: 'Aplica-se uma taxa de gestão personalizada',
        body: 'Assim que a encomenda entrar em planeamento de produção ou corte de materiais, será deduzida uma Custom Handling / Restocking Fee de 15 %-20 % pelos materiais personalizados e mão de obra já comprometidos.',
      },
      {
        id: 'shipped',
        label: 'Enviado',
        title: 'Usar o processo normal de devolução',
        body: 'Após o envio, o cancelamento direto deixa de ser aceite. Receba primeiro a encomenda e, se elegível, siga o processo normal de devolução após a receção.',
      },
    ],
  },
  ru: {
    title: 'Политика возврата средств и отмены',
    intro: 'Запросы на отмену индивидуальных заказов обрабатываются с учетом времени оплаты и этапа выполнения.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'В течение 24 часов после оплаты',
        title: 'Бесплатная отмена',
        body: 'При отмене в течение 24 часов после оплаты вы получите полный возврат 100 % на исходный способ оплаты.',
      },
      {
        id: 'after-24-hours',
        label: 'После 24 часов с момента оплаты (производство/резка запланированы)',
        title: 'Применяется сбор за индивидуальную обработку',
        body: 'После передачи заказа в планирование производства или резку материалов будет удержана Custom Handling / Restocking Fee в размере 15 %-20 % за уже зарезервированные индивидуальные материалы и трудозатраты.',
      },
      {
        id: 'shipped',
        label: 'Отправлено',
        title: 'Используйте стандартную процедуру возврата',
        body: 'После отправки прямая отмена не принимается. Сначала получите посылку, затем при соблюдении условий оформите возврат по обычной процедуре после получения.',
      },
    ],
  },
  ar: {
    title: 'سياسة الاسترداد والإلغاء',
    intro: 'تتم معالجة طلبات إلغاء الطلبات المخصصة حسب وقت الدفع ومرحلة التنفيذ.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'خلال 24 ساعة بعد الدفع',
        title: 'إلغاء مجاني',
        body: 'يمكنك الإلغاء خلال 24 ساعة بعد الدفع لاسترداد كامل بنسبة 100% إلى طريقة الدفع الأصلية.',
      },
      {
        id: 'after-24-hours',
        label: 'بعد 24 ساعة من الدفع (تمت جدولة الإنتاج/القطع)',
        title: 'تطبق رسوم معالجة مخصصة',
        body: 'بعد دخول الطلب في جدولة الإنتاج أو قطع المواد، سيتم خصم Custom Handling / Restocking Fee بنسبة 15%-20% مقابل المواد المخصصة وساعات العمل التي تم الالتزام بها بالفعل.',
      },
      {
        id: 'shipped',
        label: 'تم الشحن',
        title: 'استخدم عملية الإرجاع القياسية',
        body: 'بعد الشحن لا يتم قبول الإلغاء المباشر. يرجى استلام الطرد أولا، ثم اتباع عملية الإرجاع العادية بعد الاستلام إذا كان الطلب مؤهلا.',
      },
    ],
  },
  nl: {
    title: 'Terugbetalings- en annuleringsbeleid',
    intro: 'Annuleringsverzoeken voor maatwerkbestellingen worden behandeld op basis van het betaalmoment en de uitvoeringsfase.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Binnen 24 uur na betaling',
        title: 'Gratis annulering',
        body: 'Annuleer binnen 24 uur na betaling voor een volledige terugbetaling van 100 % naar de oorspronkelijke betaalmethode.',
      },
      {
        id: 'after-24-hours',
        label: 'Na 24 uur na betaling (ingepland voor productie/snijden)',
        title: 'Maatwerkverwerkingskosten zijn van toepassing',
        body: 'Zodra de bestelling is ingepland voor productie of het snijden van materialen, wordt een Custom Handling / Restocking Fee van 15 %-20 % ingehouden voor reeds vastgelegde maatwerkmaterialen en arbeid.',
      },
      {
        id: 'shipped',
        label: 'Verzonden',
        title: 'Gebruik het standaard retourproces',
        body: 'Na verzending wordt directe annulering niet meer geaccepteerd. Ontvang eerst het pakket en volg daarna, indien in aanmerking komend, het reguliere retourproces.',
      },
    ],
  },
  tr: {
    title: 'İade ve İptal Politikası',
    intro: 'Özel siparişler için iptal talepleri ödeme zamanı ve teslimat hazırlık aşamasına göre işleme alınır.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Ödemeden sonraki 24 saat içinde',
        title: 'Ücretsiz iptal',
        body: 'Ödemeden sonraki 24 saat içinde iptal ederseniz, orijinal ödeme yöntemine %100 tam iade yapılır.',
      },
      {
        id: 'after-24-hours',
        label: 'Ödemeden 24 saat sonra (üretim/kesim planına alındı)',
        title: 'Özel işlem ücreti uygulanır',
        body: 'Sipariş üretim planına veya malzeme kesimine girdikten sonra, ayrılmış özel malzemeler ve işçilik için %15-%20 Custom Handling / Restocking Fee kesilir.',
      },
      {
        id: 'shipped',
        label: 'Gönderildi',
        title: 'Standart iade sürecini kullanın',
        body: 'Gönderimden sonra doğrudan iptal kabul edilmez. Lütfen önce paketi teslim alın, ardından uygunsa teslimat sonrası normal iade sürecini izleyin.',
      },
    ],
  },
  id: {
    title: 'Kebijakan Pengembalian Dana & Pembatalan',
    intro: 'Permintaan pembatalan untuk pesanan kustom diproses berdasarkan waktu pembayaran dan tahap pemenuhan pesanan.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Dalam 24 jam setelah pembayaran',
        title: 'Pembatalan gratis',
        body: 'Batalkan dalam 24 jam setelah pembayaran untuk mendapatkan pengembalian dana penuh 100% ke metode pembayaran awal.',
      },
      {
        id: 'after-24-hours',
        label: 'Setelah 24 jam dari pembayaran (masuk jadwal produksi/pemotongan)',
        title: 'Biaya penanganan kustom berlaku',
        body: 'Setelah pesanan masuk ke jadwal produksi atau pemotongan material, Custom Handling / Restocking Fee sebesar 15%-20% akan dipotong untuk material kustom dan jam kerja yang sudah dialokasikan.',
      },
      {
        id: 'shipped',
        label: 'Sudah dikirim',
        title: 'Gunakan proses retur standar',
        body: 'Setelah dikirim, pembatalan langsung tidak diterima. Terima paket terlebih dahulu, lalu ikuti proses retur reguler setelah penerimaan jika memenuhi syarat.',
      },
    ],
  },
  th: {
    title: 'นโยบายการคืนเงินและการยกเลิก',
    intro: 'คำขอยกเลิกสำหรับคำสั่งซื้อแบบกำหนดเองจะพิจารณาตามเวลาชำระเงินและขั้นตอนการดำเนินงาน',
    windows: [
      {
        id: 'within-24-hours',
        label: 'ภายใน 24 ชั่วโมงหลังชำระเงิน',
        title: 'ยกเลิกฟรี',
        body: 'ยกเลิกภายใน 24 ชั่วโมงหลังชำระเงินเพื่อรับเงินคืนเต็มจำนวน 100% ไปยังวิธีชำระเงินเดิม',
      },
      {
        id: 'after-24-hours',
        label: 'หลังชำระเงินเกิน 24 ชั่วโมง (เข้าสู่การจัดตารางผลิต/ตัดวัสดุ)',
        title: 'มีค่าธรรมเนียมจัดการงานสั่งทำ',
        body: 'เมื่อคำสั่งซื้อเข้าสู่การจัดตารางผลิตหรือการตัดวัสดุแล้ว จะหัก Custom Handling / Restocking Fee 15%-20% สำหรับวัสดุสั่งทำและแรงงานที่ได้จัดสรรไปแล้ว',
      },
      {
        id: 'shipped',
        label: 'จัดส่งแล้ว',
        title: 'ใช้กระบวนการคืนสินค้าตามปกติ',
        body: 'หลังจัดส่งแล้วจะไม่รับการยกเลิกโดยตรง โปรดรับพัสดุก่อน จากนั้นหากเข้าเงื่อนไขให้ดำเนินการตามกระบวนการคืนสินค้าปกติหลังรับสินค้า',
      },
    ],
  },
  sv: {
    title: 'Policy för återbetalning och annullering',
    intro: 'Annulleringsförfrågningar för specialbeställningar hanteras utifrån betalningstidpunkt och orderns behandlingssteg.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Inom 24 timmar efter betalning',
        title: 'Kostnadsfri annullering',
        body: 'Annullera inom 24 timmar efter betalning för full återbetalning på 100 % till den ursprungliga betalningsmetoden.',
      },
      {
        id: 'after-24-hours',
        label: 'Efter 24 timmar efter betalning (planerad för produktion/kapning)',
        title: 'Avgift för specialhantering gäller',
        body: 'När ordern har gått in i produktionsplanering eller materialkapning dras en Custom Handling / Restocking Fee på 15 %-20 % för specialmaterial och arbete som redan har reserverats.',
      },
      {
        id: 'shipped',
        label: 'Skickad',
        title: 'Använd den vanliga returprocessen',
        body: 'Direkt annullering accepteras inte efter att ordern har skickats. Ta först emot paketet och följ sedan den ordinarie returprocessen om ordern är berättigad.',
      },
    ],
  },
  da: {
    title: 'Refusions- og annulleringspolitik',
    intro: 'Annulleringsanmodninger for specialordrer behandles efter betalingstidspunktet og ordrestatus.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Inden for 24 timer efter betaling',
        title: 'Gratis annullering',
        body: 'Annuller inden for 24 timer efter betaling for at få fuld refusion på 100 % til den oprindelige betalingsmetode.',
      },
      {
        id: 'after-24-hours',
        label: 'Efter 24 timer efter betaling (planlagt til produktion/skæring)',
        title: 'Gebyr for specialhåndtering gælder',
        body: 'Når ordren er gået i produktionsplanlægning eller materialeskæring, fratrækkes en Custom Handling / Restocking Fee på 15 %-20 % for specialmaterialer og arbejdstid, der allerede er afsat.',
      },
      {
        id: 'shipped',
        label: 'Afsendt',
        title: 'Brug den almindelige returproces',
        body: 'Direkte annullering accepteres ikke efter afsendelse. Modtag først pakken, og følg derefter den almindelige returproces efter modtagelsen, hvis ordren er berettiget.',
      },
    ],
  },
  fi: {
    title: 'Hyvitys- ja peruutuskäytäntö',
    intro: 'Räätälöityjen tilausten peruutuspyynnöt käsitellään maksun ajankohdan ja toimitusvaiheen perusteella.',
    windows: [
      {
        id: 'within-24-hours',
        label: '24 tunnin kuluessa maksusta',
        title: 'Maksuton peruutus',
        body: 'Peruuta 24 tunnin kuluessa maksusta saadaksesi 100 % täyden hyvityksen alkuperäiselle maksutavalle.',
      },
      {
        id: 'after-24-hours',
        label: 'Yli 24 tuntia maksun jälkeen (tuotanto/leikkaus aikataulutettu)',
        title: 'Räätälöidyn käsittelyn maksu veloitetaan',
        body: 'Kun tilaus on siirtynyt tuotannon aikataulutukseen tai materiaalien leikkaukseen, vähennetään 15 %-20 % Custom Handling / Restocking Fee jo varatuista räätälöidyistä materiaaleista ja työajasta.',
      },
      {
        id: 'shipped',
        label: 'Lähetetty',
        title: 'Käytä normaalia palautusprosessia',
        body: 'Suoraa peruutusta ei hyväksytä lähetyksen jälkeen. Vastaanota paketti ensin ja noudata sen jälkeen normaalia palautusprosessia, jos tilaus on oikeutettu palautukseen.',
      },
    ],
  },
  hi: {
    title: 'रिफंड और कैंसलेशन नीति',
    intro: 'कस्टम ऑर्डर के कैंसलेशन अनुरोध भुगतान के समय और पूर्ति चरण के आधार पर संभाले जाते हैं।',
    windows: [
      {
        id: 'within-24-hours',
        label: 'भुगतान के 24 घंटे के भीतर',
        title: 'मुफ्त कैंसलेशन',
        body: 'भुगतान के 24 घंटे के भीतर कैंसल करने पर मूल भुगतान माध्यम में 100% पूरा रिफंड मिलेगा।',
      },
      {
        id: 'after-24-hours',
        label: 'भुगतान के 24 घंटे बाद (उत्पादन/कटिंग के लिए शेड्यूल)',
        title: 'कस्टम हैंडलिंग शुल्क लागू होगा',
        body: 'ऑर्डर उत्पादन शेड्यूलिंग या सामग्री कटिंग में प्रवेश करने के बाद, पहले से लगाए गए कस्टम सामग्री और श्रम के लिए 15%-20% Custom Handling / Restocking Fee काटी जाएगी।',
      },
      {
        id: 'shipped',
        label: 'शिप हो चुका है',
        title: 'सामान्य रिटर्न प्रक्रिया अपनाएं',
        body: 'शिपमेंट के बाद सीधे कैंसलेशन स्वीकार नहीं किया जाता। कृपया पहले पैकेज प्राप्त करें, फिर पात्र होने पर प्राप्ति के बाद सामान्य रिटर्न प्रक्रिया का पालन करें।',
      },
    ],
  },
  ms: {
    title: 'Polisi Bayaran Balik & Pembatalan',
    intro: 'Permintaan pembatalan untuk pesanan tersuai diproses mengikut masa bayaran dan peringkat pemenuhan.',
    windows: [
      {
        id: 'within-24-hours',
        label: 'Dalam 24 jam selepas bayaran',
        title: 'Pembatalan percuma',
        body: 'Batalkan dalam 24 jam selepas bayaran untuk menerima bayaran balik penuh 100% kepada kaedah bayaran asal.',
      },
      {
        id: 'after-24-hours',
        label: 'Selepas 24 jam daripada bayaran (dijadualkan untuk pengeluaran/pemotongan)',
        title: 'Fi pengendalian tersuai dikenakan',
        body: 'Setelah pesanan memasuki jadual pengeluaran atau pemotongan bahan, Custom Handling / Restocking Fee sebanyak 15%-20% akan ditolak untuk bahan tersuai dan kerja yang telah diperuntukkan.',
      },
      {
        id: 'shipped',
        label: 'Telah dihantar',
        title: 'Gunakan proses pemulangan biasa',
        body: 'Selepas penghantaran, pembatalan terus tidak diterima. Sila terima bungkusan terlebih dahulu, kemudian ikut proses pemulangan biasa selepas penerimaan jika layak.',
      },
    ],
  },
}

const defaultRefundCancellationPolicyContent = (): RefundCancellationPolicyContent => {
  const content = refundCancellationPolicyByLocale.en
  if (!content) throw new Error('Missing English refund cancellation policy content.')
  return content
}

export const getRefundCancellationPolicyContent = (locale: unknown): RefundCancellationPolicyContent => {
  const localeCode = normalizeStorefrontLocaleCode(locale) || 'en'
  return refundCancellationPolicyByLocale[localeCode] || defaultRefundCancellationPolicyContent()
}
