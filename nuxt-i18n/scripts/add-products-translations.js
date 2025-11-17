import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 商品搜索结果翻译
const translations = {
  ar: { searchResults: "نتائج البحث", searchFor: "البحث عن", allProducts: "جميع المنتجات", loading: "جاري التحميل...", noResults: "لم يتم العثور على منتجات", tryAdjustFilters: "يرجى محاولة تعديل معايير التصفية", loadMore: "تحميل المزيد", items: "قطعة" },
  be: { searchResults: "Вынікі пошуку", searchFor: "Пошук", allProducts: "Усе прадукты", loading: "Загрузка...", noResults: "Прадукты не знойдзены", tryAdjustFilters: "Паспрабуйце змяніць умовы фільтрацыі", loadMore: "Загрузіць яшчэ", items: "шт" },
  bn: { searchResults: "অনুসন্ধান ফলাফল", searchFor: "অনুসন্ধান", allProducts: "সমস্ত পণ্য", loading: "লোড হচ্ছে...", noResults: "কোনো পণ্য পাওয়া যায়নি", tryAdjustFilters: "ফিল্টার সমন্বয় করার চেষ্টা করুন", loadMore: "আরও লোড করুন", items: "টি" },
  da: { searchResults: "Søgeresultater", searchFor: "Søg efter", allProducts: "Alle produkter", loading: "Indlæser...", noResults: "Ingen produkter fundet", tryAdjustFilters: "Prøv at justere filtre", loadMore: "Indlæs mere", items: "stk" },
  de: { searchResults: "Suchergebnisse", searchFor: "Suche nach", allProducts: "Alle Produkte", loading: "Lädt...", noResults: "Keine Produkte gefunden", tryAdjustFilters: "Versuchen Sie, Filter anzupassen", loadMore: "Mehr laden", items: "Stück" },
  en: { searchResults: "Search Results", searchFor: "Search for", allProducts: "All Products", loading: "Loading...", noResults: "No products found", tryAdjustFilters: "Try adjusting filters", loadMore: "Load More", items: "items" },
  es: { searchResults: "Resultados de búsqueda", searchFor: "Buscar", allProducts: "Todos los productos", loading: "Cargando...", noResults: "No se encontraron productos", tryAdjustFilters: "Intente ajustar los filtros", loadMore: "Cargar más", items: "artículos" },
  fa: { searchResults: "نتایج جستجو", searchFor: "جستجو برای", allProducts: "همه محصولات", loading: "در حال بارگذاری...", noResults: "محصولی یافت نشد", tryAdjustFilters: "فیلترها را تنظیم کنید", loadMore: "بارگذاری بیشتر", items: "مورد" },
  fi: { searchResults: "Hakutulokset", searchFor: "Hae", allProducts: "Kaikki tuotteet", loading: "Ladataan...", noResults: "Tuotteita ei löytynyt", tryAdjustFilters: "Yritä säätää suodattimia", loadMore: "Lataa lisää", items: "kpl" },
  fil: { searchResults: "Mga Resulta ng Paghahanap", searchFor: "Maghanap para sa", allProducts: "Lahat ng Produkto", loading: "Naglo-load...", noResults: "Walang nahanap na produkto", tryAdjustFilters: "Subukang ayusin ang mga filter", loadMore: "Mag-load ng Higit Pa", items: "piraso" },
  fr: { searchResults: "Résultats de recherche", searchFor: "Rechercher", allProducts: "Tous les produits", loading: "Chargement...", noResults: "Aucun produit trouvé", tryAdjustFilters: "Essayez d'ajuster les filtres", loadMore: "Charger plus", items: "articles" },
  ha: { searchResults: "Sakamakon Bincike", searchFor: "Bincika", allProducts: "Duk Kayayyaki", loading: "Ana loda...", noResults: "Ba a sami kayayyaki ba", tryAdjustFilters: "Gwada daidaita matattara", loadMore: "Loda Kari", items: "kaya" },
  hi: { searchResults: "खोज परिणाम", searchFor: "खोजें", allProducts: "सभी उत्पाद", loading: "लोड हो रहा है...", noResults: "कोई उत्पाद नहीं मिला", tryAdjustFilters: "फ़िल्टर समायोजित करने का प्रयास करें", loadMore: "और लोड करें", items: "वस्तुएं" },
  id: { searchResults: "Hasil Pencarian", searchFor: "Cari", allProducts: "Semua Produk", loading: "Memuat...", noResults: "Tidak ada produk ditemukan", tryAdjustFilters: "Coba sesuaikan filter", loadMore: "Muat Lebih Banyak", items: "item" },
  it: { searchResults: "Risultati della ricerca", searchFor: "Cerca", allProducts: "Tutti i prodotti", loading: "Caricamento...", noResults: "Nessun prodotto trovato", tryAdjustFilters: "Prova a regolare i filtri", loadMore: "Carica altro", items: "articoli" },
  ja: { searchResults: "検索結果", searchFor: "検索", allProducts: "すべての商品", loading: "読み込み中...", noResults: "商品が見つかりません", tryAdjustFilters: "フィルターを調整してください", loadMore: "もっと読み込む", items: "件" },
  jv: { searchResults: "Asil Panelusuran", searchFor: "Goleki", allProducts: "Kabeh Produk", loading: "Lagi dimuat...", noResults: "Ora ana produk ketemu", tryAdjustFilters: "Coba atur filter", loadMore: "Muat Luwih Akeh", items: "barang" },
  ko: { searchResults: "검색 결과", searchFor: "검색", allProducts: "모든 제품", loading: "로딩 중...", noResults: "제품을 찾을 수 없습니다", tryAdjustFilters: "필터를 조정해 보세요", loadMore: "더 보기", items: "개" },
  mr: { searchResults: "शोध परिणाम", searchFor: "शोधा", allProducts: "सर्व उत्पादने", loading: "लोड होत आहे...", noResults: "कोणतीही उत्पादने आढळली नाहीत", tryAdjustFilters: "फिल्टर समायोजित करण्याचा प्रयत्न करा", loadMore: "अधिक लोड करा", items: "वस्तू" },
  ms: { searchResults: "Hasil Carian", searchFor: "Cari", allProducts: "Semua Produk", loading: "Memuatkan...", noResults: "Tiada produk dijumpai", tryAdjustFilters: "Cuba laraskan penapis", loadMore: "Muat Lebih Banyak", items: "item" },
  nl: { searchResults: "Zoekresultaten", searchFor: "Zoeken naar", allProducts: "Alle producten", loading: "Laden...", noResults: "Geen producten gevonden", tryAdjustFilters: "Probeer filters aan te passen", loadMore: "Meer laden", items: "items" },
  pcm: { searchResults: "Search Results", searchFor: "Search for", allProducts: "All Products", loading: "Dey load...", noResults: "No product find", tryAdjustFilters: "Try adjust filters", loadMore: "Load More", items: "items" },
  ps: { searchResults: "د لټون پایلې", searchFor: "لټون", allProducts: "ټول محصولات", loading: "بار کیږي...", noResults: "هیڅ محصول ونه موندل شو", tryAdjustFilters: "فلټرونه تنظیم کړئ", loadMore: "نور بار کړئ", items: "توکي" },
  pt: { searchResults: "Resultados da pesquisa", searchFor: "Pesquisar", allProducts: "Todos os produtos", loading: "Carregando...", noResults: "Nenhum produto encontrado", tryAdjustFilters: "Tente ajustar os filtros", loadMore: "Carregar mais", items: "itens" },
  ru: { searchResults: "Результаты поиска", searchFor: "Поиск", allProducts: "Все товары", loading: "Загрузка...", noResults: "Товары не найдены", tryAdjustFilters: "Попробуйте изменить фильтры", loadMore: "Загрузить ещё", items: "шт" },
  sv: { searchResults: "Sökresultat", searchFor: "Sök efter", allProducts: "Alla produkter", loading: "Laddar...", noResults: "Inga produkter hittades", tryAdjustFilters: "Försök justera filter", loadMore: "Ladda mer", items: "artiklar" },
  sw: { searchResults: "Matokeo ya Utafutaji", searchFor: "Tafuta", allProducts: "Bidhaa Zote", loading: "Inapakia...", noResults: "Hakuna bidhaa zilizopatikana", tryAdjustFilters: "Jaribu kurekebisha vichujio", loadMore: "Pakia Zaidi", items: "vitu" },
  ta: { searchResults: "தேடல் முடிவுகள்", searchFor: "தேடு", allProducts: "அனைத்து தயாரிப்புகள்", loading: "ஏற்றுகிறது...", noResults: "தயாரிப்புகள் எதுவும் கிடைக்கவில்லை", tryAdjustFilters: "வடிப்பான்களை சரிசெய்ய முயற்சிக்கவும்", loadMore: "மேலும் ஏற்று", items: "பொருட்கள்" },
  te: { searchResults: "శోధన ఫలితాలు", searchFor: "శోధించండి", allProducts: "అన్ని ఉత్పత్తులు", loading: "లోడ్ అవుతోంది...", noResults: "ఉత్పత్తులు కనుగొనబడలేదు", tryAdjustFilters: "ఫిల్టర్‌లను సర్దుబాటు చేయడానికి ప్రయత్నించండి", loadMore: "మరిన్ని లోడ్ చేయండి", items: "వస్తువులు" },
  th: { searchResults: "ผลการค้นหา", searchFor: "ค้นหา", allProducts: "สินค้าทั้งหมด", loading: "กำลังโหลด...", noResults: "ไม่พบสินค้า", tryAdjustFilters: "ลองปรับตัวกรอง", loadMore: "โหลดเพิ่มเติม", items: "รายการ" },
  tl: { searchResults: "Mga Resulta ng Paghahanap", searchFor: "Maghanap para sa", allProducts: "Lahat ng Produkto", loading: "Naglo-load...", noResults: "Walang nahanap na produkto", tryAdjustFilters: "Subukang ayusin ang mga filter", loadMore: "Mag-load ng Higit Pa", items: "piraso" },
  tr: { searchResults: "Arama Sonuçları", searchFor: "Ara", allProducts: "Tüm Ürünler", loading: "Yükleniyor...", noResults: "Ürün bulunamadı", tryAdjustFilters: "Filtreleri ayarlamayı deneyin", loadMore: "Daha Fazla Yükle", items: "ürün" },
  ur: { searchResults: "تلاش کے نتائج", searchFor: "تلاش کریں", allProducts: "تمام مصنوعات", loading: "لوڈ ہو رہا ہے...", noResults: "کوئی مصنوعات نہیں ملیں", tryAdjustFilters: "فلٹرز کو ایڈجسٹ کرنے کی کوشش کریں", loadMore: "مزید لوڈ کریں", items: "اشیاء" },
  zh_cn: { searchResults: "搜索结果", searchFor: "搜索", allProducts: "所有商品", loading: "加载中...", noResults: "未找到商品", tryAdjustFilters: "请尝试调整筛选条件", loadMore: "加载更多", items: "件" }
};

const localesDir = path.join(__dirname, '../i18n/locales');

Object.keys(translations).forEach(lang => {
  const filePath = path.join(localesDir, `${lang}.json`);
  
  try {
    const content = fs.readFileSync(filePath, 'utf8');
    const json = JSON.parse(content);
    
    // 添加 products 对象
    json.products = translations[lang];
    
    fs.writeFileSync(filePath, JSON.stringify(json, null, 2), 'utf8');
    console.log(`✅ Updated ${lang}.json`);
  } catch (error) {
    console.error(`❌ Error updating ${lang}.json:`, error.message);
  }
});

console.log('\n🎉 All language files updated with products translations!');
