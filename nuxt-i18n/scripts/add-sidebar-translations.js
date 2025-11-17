import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Sidebar 翻译（所有语言）
const sidebarTranslations = {
  ar: {
    page1: "التنقل في الفئات",
    page2: "التصفية والمزيد",
    staticPages: "الصفحات",
    categories: "فئات المنتجات",
    categoriesPlaceholder: "شجرة فئات المنتجات (قيد التنفيذ)",
    quickSearch: "بحث سريع"
  },
  be: {
    page1: "Навігацыя па катэгорыях",
    page2: "Фільтр і больш",
    staticPages: "Старонкі",
    categories: "Катэгорыі прадуктаў",
    categoriesPlaceholder: "Дрэва катэгорый прадуктаў (у распрацоўцы)",
    quickSearch: "Хуткі пошук"
  },
  bn: {
    page1: "বিভাগ নেভিগেশন",
    page2: "ফিল্টার এবং আরও",
    staticPages: "পৃষ্ঠা",
    categories: "পণ্য বিভাগ",
    categoriesPlaceholder: "পণ্য বিভাগ ট্রি (বাস্তবায়ন করা হবে)",
    quickSearch: "দ্রুত অনুসন্ধান"
  },
  da: {
    page1: "Kategori Navigation",
    page2: "Filter & Mere",
    staticPages: "Sider",
    categories: "Produktkategorier",
    categoriesPlaceholder: "Produktkategoritræ (implementeres)",
    quickSearch: "Hurtig søgning"
  },
  de: {
    page1: "Kategorienavigation",
    page2: "Filter & Mehr",
    staticPages: "Seiten",
    categories: "Produktkategorien",
    categoriesPlaceholder: "Produktkategoriebaum (in Entwicklung)",
    quickSearch: "Schnellsuche"
  },
  en: {
    page1: "Category Navigation",
    page2: "Filter & More",
    staticPages: "Pages",
    categories: "Product Categories",
    categoriesPlaceholder: "Product Category Tree (To be implemented)",
    quickSearch: "Quick Search"
  },
  es: {
    page1: "Navegación de categorías",
    page2: "Filtro y más",
    staticPages: "Páginas",
    categories: "Categorías de productos",
    categoriesPlaceholder: "Árbol de categorías de productos (por implementar)",
    quickSearch: "Búsqueda rápida"
  },
  fa: {
    page1: "ناوبری دسته‌بندی",
    page2: "فیلتر و بیشتر",
    staticPages: "صفحات",
    categories: "دسته‌بندی محصولات",
    categoriesPlaceholder: "درخت دسته‌بندی محصولات (در حال پیاده‌سازی)",
    quickSearch: "جستجوی سریع"
  },
  fi: {
    page1: "Luokkanavigointi",
    page2: "Suodatin ja lisää",
    staticPages: "Sivut",
    categories: "Tuoteluokat",
    categoriesPlaceholder: "Tuoteluokkapuu (toteutetaan)",
    quickSearch: "Pikahaku"
  },
  fil: {
    page1: "Pag-navigate sa Kategorya",
    page2: "Filter at Higit Pa",
    staticPages: "Mga Pahina",
    categories: "Mga Kategorya ng Produkto",
    categoriesPlaceholder: "Puno ng Kategorya ng Produkto (Isasagawa)",
    quickSearch: "Mabilis na Paghahanap"
  },
  fr: {
    page1: "Navigation par catégorie",
    page2: "Filtre et plus",
    staticPages: "Pages",
    categories: "Catégories de produits",
    categoriesPlaceholder: "Arbre des catégories de produits (à implémenter)",
    quickSearch: "Recherche rapide"
  },
  ha: {
    page1: "Kewayawa Rukuni",
    page2: "Matattara da Kari",
    staticPages: "Shafuka",
    categories: "Rukunin Kayayyaki",
    categoriesPlaceholder: "Bishiyar Rukunin Kayayyaki (Za a aiwatar)",
    quickSearch: "Bincike Mai Sauri"
  },
  hi: {
    page1: "श्रेणी नेविगेशन",
    page2: "फ़िल्टर और अधिक",
    staticPages: "पृष्ठ",
    categories: "उत्पाद श्रेणियाँ",
    categoriesPlaceholder: "उत्पाद श्रेणी ट्री (लागू किया जाना है)",
    quickSearch: "त्वरित खोज"
  },
  id: {
    page1: "Navigasi Kategori",
    page2: "Filter & Lainnya",
    staticPages: "Halaman",
    categories: "Kategori Produk",
    categoriesPlaceholder: "Pohon Kategori Produk (Akan diimplementasikan)",
    quickSearch: "Pencarian Cepat"
  },
  it: {
    page1: "Navigazione categorie",
    page2: "Filtro e altro",
    staticPages: "Pagine",
    categories: "Categorie di prodotti",
    categoriesPlaceholder: "Albero delle categorie di prodotti (da implementare)",
    quickSearch: "Ricerca rapida"
  },
  ja: {
    page1: "カテゴリーナビゲーション",
    page2: "フィルターとその他",
    staticPages: "ページ",
    categories: "商品カテゴリー",
    categoriesPlaceholder: "商品カテゴリーツリー（実装予定）",
    quickSearch: "クイック検索"
  },
  jv: {
    page1: "Navigasi Kategori",
    page2: "Filter & Liyane",
    staticPages: "Kaca",
    categories: "Kategori Produk",
    categoriesPlaceholder: "Wit Kategori Produk (Bakal dileksanakake)",
    quickSearch: "Panelusuran Cepet"
  },
  ko: {
    page1: "카테고리 탐색",
    page2: "필터 및 더보기",
    staticPages: "페이지",
    categories: "제품 카테고리",
    categoriesPlaceholder: "제품 카테고리 트리 (구현 예정)",
    quickSearch: "빠른 검색"
  },
  mr: {
    page1: "श्रेणी नेव्हिगेशन",
    page2: "फिल्टर आणि अधिक",
    staticPages: "पृष्ठे",
    categories: "उत्पादन श्रेणी",
    categoriesPlaceholder: "उत्पादन श्रेणी वृक्ष (अंमलबजावणी करायची आहे)",
    quickSearch: "जलद शोध"
  },
  ms: {
    page1: "Navigasi Kategori",
    page2: "Penapis & Lagi",
    staticPages: "Halaman",
    categories: "Kategori Produk",
    categoriesPlaceholder: "Pokok Kategori Produk (Akan dilaksanakan)",
    quickSearch: "Carian Pantas"
  },
  nl: {
    page1: "Categorienavigatie",
    page2: "Filter & Meer",
    staticPages: "Pagina's",
    categories: "Productcategorieën",
    categoriesPlaceholder: "Productcategorieboom (te implementeren)",
    quickSearch: "Snel zoeken"
  },
  pcm: {
    page1: "Category Navigation",
    page2: "Filter & More",
    staticPages: "Pages",
    categories: "Product Categories",
    categoriesPlaceholder: "Product Category Tree (Go implement am)",
    quickSearch: "Quick Search"
  },
  ps: {
    page1: "د کټګورۍ نیویګیشن",
    page2: "فلټر او نور",
    staticPages: "پاڼې",
    categories: "د محصول کټګورۍ",
    categoriesPlaceholder: "د محصول کټګورۍ ونه (تطبیق کیږي)",
    quickSearch: "ګړندۍ لټون"
  },
  pt: {
    page1: "Navegação de categorias",
    page2: "Filtro e mais",
    staticPages: "Páginas",
    categories: "Categorias de produtos",
    categoriesPlaceholder: "Árvore de categorias de produtos (a ser implementado)",
    quickSearch: "Pesquisa rápida"
  },
  ru: {
    page1: "Навигация по категориям",
    page2: "Фильтр и ещё",
    staticPages: "Страницы",
    categories: "Категории товаров",
    categoriesPlaceholder: "Дерево категорий товаров (в разработке)",
    quickSearch: "Быстрый поиск"
  },
  sv: {
    page1: "Kategorinavigering",
    page2: "Filter & Mer",
    staticPages: "Sidor",
    categories: "Produktkategorier",
    categoriesPlaceholder: "Produktkategoriträd (att implementeras)",
    quickSearch: "Snabbsökning"
  },
  sw: {
    page1: "Urambazaji wa Kategoria",
    page2: "Chuja na Zaidi",
    staticPages: "Kurasa",
    categories: "Kategoria za Bidhaa",
    categoriesPlaceholder: "Mti wa Kategoria za Bidhaa (Utatekelezwa)",
    quickSearch: "Utafutaji wa Haraka"
  },
  ta: {
    page1: "வகை வழிசெலுத்தல்",
    page2: "வடிப்பான் மற்றும் மேலும்",
    staticPages: "பக்கங்கள்",
    categories: "தயாரிப்பு வகைகள்",
    categoriesPlaceholder: "தயாரிப்பு வகை மரம் (செயல்படுத்தப்படும்)",
    quickSearch: "விரைவு தேடல்"
  },
  te: {
    page1: "వర్గం నావిగేషన్",
    page2: "ఫిల్టర్ & మరిన్ని",
    staticPages: "పేజీలు",
    categories: "ఉత్పత్తి వర్గాలు",
    categoriesPlaceholder: "ఉత్పత్తి వర్గం ట్రీ (అమలు చేయబడుతుంది)",
    quickSearch: "త్వరిత శోధన"
  },
  th: {
    page1: "การนำทางหมวดหมู่",
    page2: "ตัวกรองและอื่นๆ",
    staticPages: "หน้า",
    categories: "หมวดหมู่สินค้า",
    categoriesPlaceholder: "ต้นไม้หมวดหมู่สินค้า (จะดำเนินการ)",
    quickSearch: "ค้นหาด่วน"
  },
  tl: {
    page1: "Pag-navigate sa Kategorya",
    page2: "Filter at Higit Pa",
    staticPages: "Mga Pahina",
    categories: "Mga Kategorya ng Produkto",
    categoriesPlaceholder: "Puno ng Kategorya ng Produkto (Isasagawa)",
    quickSearch: "Mabilis na Paghahanap"
  },
  tr: {
    page1: "Kategori Gezinme",
    page2: "Filtre ve Daha Fazlası",
    staticPages: "Sayfalar",
    categories: "Ürün Kategorileri",
    categoriesPlaceholder: "Ürün Kategori Ağacı (uygulanacak)",
    quickSearch: "Hızlı Arama"
  },
  ur: {
    page1: "زمرہ نیویگیشن",
    page2: "فلٹر اور مزید",
    staticPages: "صفحات",
    categories: "مصنوعات کی اقسام",
    categoriesPlaceholder: "مصنوعات کی اقسام کا درخت (نافذ کیا جائے گا)",
    quickSearch: "فوری تلاش"
  },
  zh_cn: {
    page1: "分类导航",
    page2: "筛选 & 更多",
    staticPages: "页面",
    categories: "商品分类",
    categoriesPlaceholder: "商品分类树（待实现）",
    quickSearch: "快速搜索"
  }
};

const localesDir = path.join(__dirname, '../i18n/locales');

// 读取所有语言文件并添加 sidebar 翻译
Object.keys(sidebarTranslations).forEach(lang => {
  const filePath = path.join(localesDir, `${lang}.json`);
  
  try {
    // 读取现有文件
    const content = fs.readFileSync(filePath, 'utf8');
    const json = JSON.parse(content);
    
    // 添加 sidebar 翻译
    json.sidebar = sidebarTranslations[lang];
    
    // 写回文件
    fs.writeFileSync(filePath, JSON.stringify(json, null, 2), 'utf8');
    console.log(`✅ Updated ${lang}.json`);
  } catch (error) {
    console.error(`❌ Error updating ${lang}.json:`, error.message);
  }
});

console.log('\n🎉 All language files updated with sidebar translations!');
