import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 筛选器翻译（所有语言）
const filterTranslations = {
  ar: {
    title: "مرشحات متقدمة",
    priceRange: "نطاق السعر",
    stockStatus: "حالة المخزون",
    sortBy: "ترتيب حسب",
    rating: "التقييم",
    inStock: "متوفر",
    preOrder: "طلب مسبق",
    andUp: "وما فوق",
    reset: "إعادة تعيين المرشحات",
    sort: {
      newest: "الأحدث",
      priceLowToHigh: "السعر: من الأقل إلى الأعلى",
      priceHighToLow: "السعر: من الأعلى إلى الأقل",
      popular: "الأكثر شعبية",
      rating: "أفضل تقييم"
    }
  },
  be: {
    title: "Пашыраны фільтр",
    priceRange: "Дыяпазон цэн",
    stockStatus: "Статус запасаў",
    sortBy: "Сартаваць па",
    rating: "Рэйтынг",
    inStock: "У наяўнасці",
    preOrder: "Папярэдні заказ",
    andUp: "і вышэй",
    reset: "Скінуць фільтры",
    sort: {
      newest: "Найноўшыя",
      priceLowToHigh: "Цана: ад нізкай да высокай",
      priceHighToLow: "Цана: ад высокай да нізкай",
      popular: "Самыя папулярныя",
      rating: "Лепшы рэйтынг"
    }
  },
  bn: {
    title: "উন্নত ফিল্টার",
    priceRange: "মূল্য পরিসীমা",
    stockStatus: "স্টক স্ট্যাটাস",
    sortBy: "সাজান",
    rating: "রেটিং",
    inStock: "স্টকে আছে",
    preOrder: "প্রি-অর্ডার",
    andUp: "এবং উপরে",
    reset: "ফিল্টার রিসেট করুন",
    sort: {
      newest: "নতুন",
      priceLowToHigh: "মূল্য: কম থেকে বেশি",
      priceHighToLow: "মূল্য: বেশি থেকে কম",
      popular: "সবচেয়ে জনপ্রিয়",
      rating: "সেরা রেটিং"
    }
  },
  da: {
    title: "Avancerede filtre",
    priceRange: "Prisinterval",
    stockStatus: "Lagerstatus",
    sortBy: "Sortér efter",
    rating: "Bedømmelse",
    inStock: "På lager",
    preOrder: "Forudbestilling",
    andUp: "og op",
    reset: "Nulstil filtre",
    sort: {
      newest: "Nyeste",
      priceLowToHigh: "Pris: Lav til høj",
      priceHighToLow: "Pris: Høj til lav",
      popular: "Mest populære",
      rating: "Bedste bedømmelse"
    }
  },
  de: {
    title: "Erweiterte Filter",
    priceRange: "Preisspanne",
    stockStatus: "Lagerstatus",
    sortBy: "Sortieren nach",
    rating: "Bewertung",
    inStock: "Auf Lager",
    preOrder: "Vorbestellung",
    andUp: "und höher",
    reset: "Filter zurücksetzen",
    sort: {
      newest: "Neueste",
      priceLowToHigh: "Preis: Niedrig bis Hoch",
      priceHighToLow: "Preis: Hoch bis Niedrig",
      popular: "Am beliebtesten",
      rating: "Beste Bewertung"
    }
  },
  en: {
    title: "Advanced Filters",
    priceRange: "Price Range",
    stockStatus: "Stock Status",
    sortBy: "Sort By",
    rating: "Rating",
    inStock: "In Stock",
    preOrder: "Pre-order",
    andUp: "& Up",
    reset: "Reset Filters",
    sort: {
      newest: "Newest",
      priceLowToHigh: "Price: Low to High",
      priceHighToLow: "Price: High to Low",
      popular: "Most Popular",
      rating: "Best Rating"
    }
  },
  es: {
    title: "Filtros avanzados",
    priceRange: "Rango de precios",
    stockStatus: "Estado de stock",
    sortBy: "Ordenar por",
    rating: "Calificación",
    inStock: "En stock",
    preOrder: "Pre-pedido",
    andUp: "y más",
    reset: "Restablecer filtros",
    sort: {
      newest: "Más reciente",
      priceLowToHigh: "Precio: Bajo a Alto",
      priceHighToLow: "Precio: Alto a Bajo",
      popular: "Más popular",
      rating: "Mejor calificación"
    }
  },
  fa: {
    title: "فیلترهای پیشرفته",
    priceRange: "محدوده قیمت",
    stockStatus: "وضعیت موجودی",
    sortBy: "مرتب سازی بر اساس",
    rating: "امتیاز",
    inStock: "موجود",
    preOrder: "پیش سفارش",
    andUp: "و بالاتر",
    reset: "بازنشانی فیلترها",
    sort: {
      newest: "جدیدترین",
      priceLowToHigh: "قیمت: کم به زیاد",
      priceHighToLow: "قیمت: زیاد به کم",
      popular: "محبوب ترین",
      rating: "بهترین امتیاز"
    }
  },
  fi: {
    title: "Lisäsuodattimet",
    priceRange: "Hintahaarukka",
    stockStatus: "Varastotilanne",
    sortBy: "Lajittele",
    rating: "Arvostelu",
    inStock: "Varastossa",
    preOrder: "Ennakkotilaus",
    andUp: "ja ylöspäin",
    reset: "Nollaa suodattimet",
    sort: {
      newest: "Uusin",
      priceLowToHigh: "Hinta: Matala - Korkea",
      priceHighToLow: "Hinta: Korkea - Matala",
      popular: "Suosituin",
      rating: "Paras arvostelu"
    }
  },
  fil: {
    title: "Advanced na Mga Filter",
    priceRange: "Hanay ng Presyo",
    stockStatus: "Katayuan ng Stock",
    sortBy: "Ayusin Ayon sa",
    rating: "Rating",
    inStock: "May Stock",
    preOrder: "Pre-order",
    andUp: "at Pataas",
    reset: "I-reset ang Mga Filter",
    sort: {
      newest: "Pinakabago",
      priceLowToHigh: "Presyo: Mababa hanggang Mataas",
      priceHighToLow: "Presyo: Mataas hanggang Mababa",
      popular: "Pinakasikat",
      rating: "Pinakamahusay na Rating"
    }
  },
  fr: {
    title: "Filtres avancés",
    priceRange: "Gamme de prix",
    stockStatus: "État du stock",
    sortBy: "Trier par",
    rating: "Évaluation",
    inStock: "En stock",
    preOrder: "Précommande",
    andUp: "et plus",
    reset: "Réinitialiser les filtres",
    sort: {
      newest: "Plus récent",
      priceLowToHigh: "Prix: Bas à Élevé",
      priceHighToLow: "Prix: Élevé à Bas",
      popular: "Plus populaire",
      rating: "Meilleure évaluation"
    }
  },
  ha: {
    title: "Matattara Matattara",
    priceRange: "Kewayon Farashi",
    stockStatus: "Matsayin Kaya",
    sortBy: "Tsara Ta",
    rating: "Kimanta",
    inStock: "A Cikin Kaya",
    preOrder: "Oda Ta Gaba",
    andUp: "da Sama",
    reset: "Sake Saita Matattara",
    sort: {
      newest: "Sabuwa",
      priceLowToHigh: "Farashi: Ƙasa zuwa Sama",
      priceHighToLow: "Farashi: Sama zuwa Ƙasa",
      popular: "Mafi Shahara",
      rating: "Mafi Kyawun Kimanta"
    }
  },
  hi: {
    title: "उन्नत फ़िल्टर",
    priceRange: "मूल्य सीमा",
    stockStatus: "स्टॉक स्थिति",
    sortBy: "इसके अनुसार क्रमबद्ध करें",
    rating: "रेटिंग",
    inStock: "स्टॉक में",
    preOrder: "प्री-ऑर्डर",
    andUp: "और ऊपर",
    reset: "फ़िल्टर रीसेट करें",
    sort: {
      newest: "नवीनतम",
      priceLowToHigh: "मूल्य: कम से अधिक",
      priceHighToLow: "मूल्य: अधिक से कम",
      popular: "सबसे लोकप्रिय",
      rating: "सर्वश्रेष्ठ रेटिंग"
    }
  },
  id: {
    title: "Filter Lanjutan",
    priceRange: "Rentang Harga",
    stockStatus: "Status Stok",
    sortBy: "Urutkan Berdasarkan",
    rating: "Penilaian",
    inStock: "Tersedia",
    preOrder: "Pre-order",
    andUp: "& Ke Atas",
    reset: "Reset Filter",
    sort: {
      newest: "Terbaru",
      priceLowToHigh: "Harga: Rendah ke Tinggi",
      priceHighToLow: "Harga: Tinggi ke Rendah",
      popular: "Paling Populer",
      rating: "Penilaian Terbaik"
    }
  },
  it: {
    title: "Filtri avanzati",
    priceRange: "Fascia di prezzo",
    stockStatus: "Stato delle scorte",
    sortBy: "Ordina per",
    rating: "Valutazione",
    inStock: "Disponibile",
    preOrder: "Pre-ordine",
    andUp: "e oltre",
    reset: "Reimposta filtri",
    sort: {
      newest: "Più recente",
      priceLowToHigh: "Prezzo: Basso ad Alto",
      priceHighToLow: "Prezzo: Alto a Basso",
      popular: "Più popolare",
      rating: "Migliore valutazione"
    }
  },
  ja: {
    title: "詳細フィルター",
    priceRange: "価格帯",
    stockStatus: "在庫状況",
    sortBy: "並び替え",
    rating: "評価",
    inStock: "在庫あり",
    preOrder: "予約注文",
    andUp: "以上",
    reset: "フィルターをリセット",
    sort: {
      newest: "新着順",
      priceLowToHigh: "価格: 安い順",
      priceHighToLow: "価格: 高い順",
      popular: "人気順",
      rating: "評価順"
    }
  },
  jv: {
    title: "Filter Lanjut",
    priceRange: "Rentang Rega",
    stockStatus: "Status Stok",
    sortBy: "Urutake Miturut",
    rating: "Rating",
    inStock: "Ana Stok",
    preOrder: "Pre-order",
    andUp: "lan Munggah",
    reset: "Reset Filter",
    sort: {
      newest: "Paling Anyar",
      priceLowToHigh: "Rega: Murah menyang Larang",
      priceHighToLow: "Rega: Larang menyang Murah",
      popular: "Paling Populer",
      rating: "Rating Paling Apik"
    }
  },
  ko: {
    title: "고급 필터",
    priceRange: "가격 범위",
    stockStatus: "재고 상태",
    sortBy: "정렬 기준",
    rating: "평점",
    inStock: "재고 있음",
    preOrder: "예약 주문",
    andUp: "이상",
    reset: "필터 초기화",
    sort: {
      newest: "최신순",
      priceLowToHigh: "가격: 낮은순",
      priceHighToLow: "가격: 높은순",
      popular: "인기순",
      rating: "평점순"
    }
  },
  mr: {
    title: "प्रगत फिल्टर",
    priceRange: "किंमत श्रेणी",
    stockStatus: "स्टॉक स्थिती",
    sortBy: "यानुसार क्रमवारी लावा",
    rating: "रेटिंग",
    inStock: "स्टॉकमध्ये",
    preOrder: "प्री-ऑर्डर",
    andUp: "आणि वर",
    reset: "फिल्टर रीसेट करा",
    sort: {
      newest: "नवीनतम",
      priceLowToHigh: "किंमत: कमी ते जास्त",
      priceHighToLow: "किंमत: जास्त ते कमी",
      popular: "सर्वात लोकप्रिय",
      rating: "सर्वोत्तम रेटिंग"
    }
  },
  ms: {
    title: "Penapis Lanjutan",
    priceRange: "Julat Harga",
    stockStatus: "Status Stok",
    sortBy: "Isih Mengikut",
    rating: "Penilaian",
    inStock: "Ada Stok",
    preOrder: "Pra-tempahan",
    andUp: "& Ke Atas",
    reset: "Set Semula Penapis",
    sort: {
      newest: "Terbaru",
      priceLowToHigh: "Harga: Rendah ke Tinggi",
      priceHighToLow: "Harga: Tinggi ke Rendah",
      popular: "Paling Popular",
      rating: "Penilaian Terbaik"
    }
  },
  nl: {
    title: "Geavanceerde filters",
    priceRange: "Prijsbereik",
    stockStatus: "Voorraadstatus",
    sortBy: "Sorteer op",
    rating: "Beoordeling",
    inStock: "Op voorraad",
    preOrder: "Voorbestelling",
    andUp: "en hoger",
    reset: "Filters resetten",
    sort: {
      newest: "Nieuwste",
      priceLowToHigh: "Prijs: Laag naar Hoog",
      priceHighToLow: "Prijs: Hoog naar Laag",
      popular: "Meest populair",
      rating: "Beste beoordeling"
    }
  },
  pcm: {
    title: "Advanced Filters",
    priceRange: "Price Range",
    stockStatus: "Stock Status",
    sortBy: "Sort By",
    rating: "Rating",
    inStock: "Dey for Stock",
    preOrder: "Pre-order",
    andUp: "& Up",
    reset: "Reset Filters",
    sort: {
      newest: "Latest",
      priceLowToHigh: "Price: Low to High",
      priceHighToLow: "Price: High to Low",
      popular: "Most Popular",
      rating: "Best Rating"
    }
  },
  ps: {
    title: "پرمختللي فلټرونه",
    priceRange: "د قیمت حد",
    stockStatus: "د ذخیرې حالت",
    sortBy: "ترتیب کول",
    rating: "درجه بندي",
    inStock: "په ذخیره کې",
    preOrder: "مخکینۍ امر",
    andUp: "او پورته",
    reset: "فلټرونه بیا تنظیم کړئ",
    sort: {
      newest: "تازه",
      priceLowToHigh: "قیمت: ټیټ څخه لوړ",
      priceHighToLow: "قیمت: لوړ څخه ټیټ",
      popular: "خورا مشهور",
      rating: "غوره درجه بندي"
    }
  },
  pt: {
    title: "Filtros avançados",
    priceRange: "Faixa de preço",
    stockStatus: "Status do estoque",
    sortBy: "Ordenar por",
    rating: "Avaliação",
    inStock: "Em estoque",
    preOrder: "Pré-venda",
    andUp: "e acima",
    reset: "Redefinir filtros",
    sort: {
      newest: "Mais recente",
      priceLowToHigh: "Preço: Baixo para Alto",
      priceHighToLow: "Preço: Alto para Baixo",
      popular: "Mais popular",
      rating: "Melhor avaliação"
    }
  },
  ru: {
    title: "Расширенные фильтры",
    priceRange: "Диапазон цен",
    stockStatus: "Статус наличия",
    sortBy: "Сортировать по",
    rating: "Рейтинг",
    inStock: "В наличии",
    preOrder: "Предзаказ",
    andUp: "и выше",
    reset: "Сбросить фильтры",
    sort: {
      newest: "Новинки",
      priceLowToHigh: "Цена: По возрастанию",
      priceHighToLow: "Цена: По убыванию",
      popular: "Популярные",
      rating: "Лучший рейтинг"
    }
  },
  sv: {
    title: "Avancerade filter",
    priceRange: "Prisintervall",
    stockStatus: "Lagerstatus",
    sortBy: "Sortera efter",
    rating: "Betyg",
    inStock: "I lager",
    preOrder: "Förbeställning",
    andUp: "och uppåt",
    reset: "Återställ filter",
    sort: {
      newest: "Nyaste",
      priceLowToHigh: "Pris: Låg till Hög",
      priceHighToLow: "Pris: Hög till Låg",
      popular: "Mest populära",
      rating: "Bästa betyg"
    }
  },
  sw: {
    title: "Vichujio vya Juu",
    priceRange: "Kiwango cha Bei",
    stockStatus: "Hali ya Hifadhi",
    sortBy: "Panga Kwa",
    rating: "Ukadiriaji",
    inStock: "Ipo Hifadhini",
    preOrder: "Agiza Mapema",
    andUp: "na Zaidi",
    reset: "Weka Upya Vichujio",
    sort: {
      newest: "Mpya Zaidi",
      priceLowToHigh: "Bei: Chini hadi Juu",
      priceHighToLow: "Bei: Juu hadi Chini",
      popular: "Maarufu Zaidi",
      rating: "Ukadiriaji Bora"
    }
  },
  ta: {
    title: "மேம்பட்ட வடிப்பான்கள்",
    priceRange: "விலை வரம்பு",
    stockStatus: "இருப்பு நிலை",
    sortBy: "வரிசைப்படுத்து",
    rating: "மதிப்பீடு",
    inStock: "கையிருப்பில் உள்ளது",
    preOrder: "முன்பதிவு",
    andUp: "மற்றும் மேல்",
    reset: "வடிப்பான்களை மீட்டமை",
    sort: {
      newest: "புதியவை",
      priceLowToHigh: "விலை: குறைவு முதல் அதிகம்",
      priceHighToLow: "விலை: அதிகம் முதல் குறைவு",
      popular: "மிகவும் பிரபலமானவை",
      rating: "சிறந்த மதிப்பீடு"
    }
  },
  te: {
    title: "అధునాతన ఫిల్టర్‌లు",
    priceRange: "ధర పరిధి",
    stockStatus: "స్టాక్ స్థితి",
    sortBy: "ఇలా క్రమబద్ధీకరించు",
    rating: "రేటింగ్",
    inStock: "స్టాక్‌లో ఉంది",
    preOrder: "ముందస్తు ఆర్డర్",
    andUp: "మరియు పైన",
    reset: "ఫిల్టర్‌లను రీసెట్ చేయండి",
    sort: {
      newest: "కొత్తవి",
      priceLowToHigh: "ధర: తక్కువ నుండి ఎక్కువ",
      priceHighToLow: "ధర: ఎక్కువ నుండి తక్కువ",
      popular: "అత్యంత ప్రజాదరణ పొందినవి",
      rating: "ఉత్తమ రేటింగ్"
    }
  },
  th: {
    title: "ตัวกรองขั้นสูง",
    priceRange: "ช่วงราคา",
    stockStatus: "สถานะสต็อก",
    sortBy: "เรียงตาม",
    rating: "คะแนน",
    inStock: "มีสินค้า",
    preOrder: "สั่งจองล่วงหน้า",
    andUp: "ขึ้นไป",
    reset: "รีเซ็ตตัวกรอง",
    sort: {
      newest: "ใหม่ล่าสุด",
      priceLowToHigh: "ราคา: ต่ำไปสูง",
      priceHighToLow: "ราคา: สูงไปต่ำ",
      popular: "ยอดนิยมสูงสุด",
      rating: "คะแนนดีที่สุด"
    }
  },
  tl: {
    title: "Mga Advanced na Filter",
    priceRange: "Saklaw ng Presyo",
    stockStatus: "Katayuan ng Stock",
    sortBy: "Pagbukud-bukurin Ayon sa",
    rating: "Rating",
    inStock: "May Stock",
    preOrder: "Pre-order",
    andUp: "at Pataas",
    reset: "I-reset ang Mga Filter",
    sort: {
      newest: "Pinakabago",
      priceLowToHigh: "Presyo: Mababa hanggang Mataas",
      priceHighToLow: "Presyo: Mataas hanggang Mababa",
      popular: "Pinakasikat",
      rating: "Pinakamahusay na Rating"
    }
  },
  tr: {
    title: "Gelişmiş Filtreler",
    priceRange: "Fiyat Aralığı",
    stockStatus: "Stok Durumu",
    sortBy: "Sırala",
    rating: "Değerlendirme",
    inStock: "Stokta",
    preOrder: "Ön Sipariş",
    andUp: "ve Üzeri",
    reset: "Filtreleri Sıfırla",
    sort: {
      newest: "En Yeni",
      priceLowToHigh: "Fiyat: Düşükten Yükseğe",
      priceHighToLow: "Fiyat: Yüksekten Düşüğe",
      popular: "En Popüler",
      rating: "En İyi Değerlendirme"
    }
  },
  ur: {
    title: "جدید فلٹرز",
    priceRange: "قیمت کی حد",
    stockStatus: "اسٹاک کی حیثیت",
    sortBy: "ترتیب دیں",
    rating: "درجہ بندی",
    inStock: "اسٹاک میں",
    preOrder: "پہلے سے آرڈر",
    andUp: "اور اوپر",
    reset: "فلٹرز دوبارہ ترتیب دیں",
    sort: {
      newest: "تازہ ترین",
      priceLowToHigh: "قیمت: کم سے زیادہ",
      priceHighToLow: "قیمت: زیادہ سے کم",
      popular: "سب سے مقبول",
      rating: "بہترین درجہ بندی"
    }
  },
  zh_cn: {
    title: "高级筛选",
    priceRange: "价格范围",
    stockStatus: "库存状态",
    sortBy: "排序方式",
    rating: "评分",
    inStock: "有货",
    preOrder: "预售",
    andUp: "及以上",
    reset: "重置筛选",
    sort: {
      newest: "最新",
      priceLowToHigh: "价格从低到高",
      priceHighToLow: "价格从高到低",
      popular: "最受欢迎",
      rating: "评分最高"
    }
  }
};

const localesDir = path.join(__dirname, '../i18n/locales');

// 读取所有语言文件并添加筛选器翻译
Object.keys(filterTranslations).forEach(lang => {
  const filePath = path.join(localesDir, `${lang}.json`);
  
  try {
    // 读取现有文件
    const content = fs.readFileSync(filePath, 'utf8');
    const json = JSON.parse(content);
    
    // 添加 filter 翻译
    json.filter = filterTranslations[lang];
    
    // 写回文件
    fs.writeFileSync(filePath, JSON.stringify(json, null, 2), 'utf8');
    console.log(`✅ Updated ${lang}.json`);
  } catch (error) {
    console.error(`❌ Error updating ${lang}.json:`, error.message);
  }
});

console.log('\n🎉 All language files updated!');
