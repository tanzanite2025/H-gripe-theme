import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 商品搜索翻译
const translations = {
  ar: { productSearch: "البحث عن المنتجات", searchProductPlaceholder: "أدخل اسم المنتج...", searchProducts: "البحث عن المنتجات" },
  be: { productSearch: "Пошук прадуктаў", searchProductPlaceholder: "Увядзіце назву прадукту...", searchProducts: "Шукаць прадукты" },
  bn: { productSearch: "পণ্য অনুসন্ধান", searchProductPlaceholder: "পণ্যের নাম লিখুন...", searchProducts: "পণ্য খুঁজুন" },
  da: { productSearch: "Produktsøgning", searchProductPlaceholder: "Indtast produktnavn...", searchProducts: "Søg produkter" },
  de: { productSearch: "Produktsuche", searchProductPlaceholder: "Produktname eingeben...", searchProducts: "Produkte suchen" },
  en: { productSearch: "Product Search", searchProductPlaceholder: "Enter product name...", searchProducts: "Search Products" },
  es: { productSearch: "Búsqueda de productos", searchProductPlaceholder: "Ingrese nombre del producto...", searchProducts: "Buscar productos" },
  fa: { productSearch: "جستجوی محصول", searchProductPlaceholder: "نام محصول را وارد کنید...", searchProducts: "جستجوی محصولات" },
  fi: { productSearch: "Tuotehaku", searchProductPlaceholder: "Syötä tuotteen nimi...", searchProducts: "Hae tuotteita" },
  fil: { productSearch: "Paghahanap ng Produkto", searchProductPlaceholder: "Ilagay ang pangalan ng produkto...", searchProducts: "Maghanap ng Produkto" },
  fr: { productSearch: "Recherche de produits", searchProductPlaceholder: "Entrez le nom du produit...", searchProducts: "Rechercher des produits" },
  ha: { productSearch: "Binciken Kayayyaki", searchProductPlaceholder: "Shigar da sunan kayayyaki...", searchProducts: "Bincika Kayayyaki" },
  hi: { productSearch: "उत्पाद खोज", searchProductPlaceholder: "उत्पाद का नाम दर्ज करें...", searchProducts: "उत्पाद खोजें" },
  id: { productSearch: "Pencarian Produk", searchProductPlaceholder: "Masukkan nama produk...", searchProducts: "Cari Produk" },
  it: { productSearch: "Ricerca prodotti", searchProductPlaceholder: "Inserisci nome prodotto...", searchProducts: "Cerca prodotti" },
  ja: { productSearch: "商品検索", searchProductPlaceholder: "商品名を入力...", searchProducts: "商品を検索" },
  jv: { productSearch: "Panelusuran Produk", searchProductPlaceholder: "Lebokake jeneng produk...", searchProducts: "Goleki Produk" },
  ko: { productSearch: "제품 검색", searchProductPlaceholder: "제품명 입력...", searchProducts: "제품 검색" },
  mr: { productSearch: "उत्पादन शोध", searchProductPlaceholder: "उत्पादनाचे नाव प्रविष्ट करा...", searchProducts: "उत्पादने शोधा" },
  ms: { productSearch: "Carian Produk", searchProductPlaceholder: "Masukkan nama produk...", searchProducts: "Cari Produk" },
  nl: { productSearch: "Producten zoeken", searchProductPlaceholder: "Voer productnaam in...", searchProducts: "Zoek producten" },
  pcm: { productSearch: "Product Search", searchProductPlaceholder: "Enter product name...", searchProducts: "Search Products" },
  ps: { productSearch: "د محصول لټون", searchProductPlaceholder: "د محصول نوم دننه کړئ...", searchProducts: "محصولات ولټوئ" },
  pt: { productSearch: "Pesquisa de produtos", searchProductPlaceholder: "Digite o nome do produto...", searchProducts: "Pesquisar produtos" },
  ru: { productSearch: "Поиск товаров", searchProductPlaceholder: "Введите название товара...", searchProducts: "Искать товары" },
  sv: { productSearch: "Produktsökning", searchProductPlaceholder: "Ange produktnamn...", searchProducts: "Sök produkter" },
  sw: { productSearch: "Utafutaji wa Bidhaa", searchProductPlaceholder: "Weka jina la bidhaa...", searchProducts: "Tafuta Bidhaa" },
  ta: { productSearch: "தயாரிப்பு தேடல்", searchProductPlaceholder: "தயாரிப்பு பெயரை உள்ளிடவும்...", searchProducts: "தயாரிப்புகளைத் தேடு" },
  te: { productSearch: "ఉత్పత్తి శోధన", searchProductPlaceholder: "ఉత్పత్తి పేరు నమోదు చేయండి...", searchProducts: "ఉత్పత్తులను శోధించండి" },
  th: { productSearch: "ค้นหาสินค้า", searchProductPlaceholder: "ป้อนชื่อสินค้า...", searchProducts: "ค้นหาสินค้า" },
  tl: { productSearch: "Paghahanap ng Produkto", searchProductPlaceholder: "Ilagay ang pangalan ng produkto...", searchProducts: "Maghanap ng Produkto" },
  tr: { productSearch: "Ürün Arama", searchProductPlaceholder: "Ürün adını girin...", searchProducts: "Ürün Ara" },
  ur: { productSearch: "مصنوعات کی تلاش", searchProductPlaceholder: "مصنوعات کا نام درج کریں...", searchProducts: "مصنوعات تلاش کریں" },
  zh_cn: { productSearch: "商品搜索", searchProductPlaceholder: "输入商品名称...", searchProducts: "搜索商品" }
};

const localesDir = path.join(__dirname, '../i18n/locales');

Object.keys(translations).forEach(lang => {
  const filePath = path.join(localesDir, `${lang}.json`);
  
  try {
    const content = fs.readFileSync(filePath, 'utf8');
    const json = JSON.parse(content);
    
    // 添加到 sidebar 对象
    if (!json.sidebar) json.sidebar = {};
    Object.assign(json.sidebar, translations[lang as keyof typeof translations]);
    
    fs.writeFileSync(filePath, JSON.stringify(json, null, 2), 'utf8');
    console.log(`✅ Updated ${lang}.json`);
  } catch (error: unknown) {
    console.error(`❌ Error updating ${lang}.json:`, error instanceof Error ? error.message : error);
  }
});

console.log('\n🎉 All language files updated with product search translations!');
