const fs = require('fs');
const path = require('path');

const localesDir = path.join(__dirname, 'itinera-web', 'src', 'lib', 'i18n');
const files = ['en.json', 'es.json', 'jp.json', 'id.json'];

const translations = {
  en: {
    roma: { title: "Roma Express: 3 unmissable days", desc: "Optimized itinerary for a short city break. Colosseum, Vatican, Trastevere, and local gastronomy. Designed for travelers who value time." },
    japan: { title: "Classic Japan: 15 days from Tokyo to Hiroshima", desc: "Optimized route with 7-day JR Pass. Balance between culture, nature, and gastronomy. Includes basic etiquette and useful phrases." },
    asia: { title: "Deep Asia: 60 days from Bangkok to Bali", desc: "Route for digital nomads and experienced backpackers. Tight budget, coworkings, and local connections. Multi-currency active." }
  },
  es: {
    roma: { title: "Roma Express: 3 días imperdibles", desc: "Itinerario optimizado para un city break corto. Coliseo, Vaticano, Trastevere y gastronomía local. Diseñado para viajeros que valoran el tiempo." },
    japan: { title: "Japón Clásico: 15 días de Tokio a Hiroshima", desc: "Ruta optimizada con JR Pass 7-day. Equilibrio entre cultura, naturaleza y gastronomía. Incluye etiqueta básica y frases útiles." },
    asia: { title: "Asia Profunda: 60 días de Bangkok a Bali", desc: "Ruta para nómadas digitales y mochileros experimentados. Presupuesto ajustado, coworkings, y conexiones locales. Multi-moneda activa." }
  },
  jp: {
    roma: { title: "ローマ・エクスプレス：見逃せない3日間", desc: "短いシティブレイクに最適な旅程。コロッセオ、バチカン、トラステヴェレ、地元グルメ。時間を大切にする旅行者向け。" },
    japan: { title: "クラシック・ジャパン：東京から広島までの15日間", desc: "7日間のJRパスを活用した最適ルート。文化、自然、グルメのバランス。基本的なマナーと便利なフレーズ付き。" },
    asia: { title: "ディープ・アジア：バンコクからバリまでの60日間", desc: "デジタルノマドや経験豊富なバックパッカー向けのルート。限られた予算、コワーキング、現地でのつながり。多通貨対応。" }
  },
  id: {
    roma: { title: "Roma Express: 3 hari yang tak terlupakan", desc: "Rencana perjalanan yang dioptimalkan untuk liburan singkat di kota. Colosseum, Vatikan, Trastevere, dan keahlian memasak lokal. Dirancang untuk wisatawan yang menghargai waktu." },
    japan: { title: "Jepang Klasik: 15 hari dari Tokyo ke Hiroshima", desc: "Rute yang dioptimalkan dengan JR Pass 7-hari. Keseimbangan antara budaya, alam, dan gastronomi. Termasuk etiket dasar dan frasa berguna." },
    asia: { title: "Asia Mendalam: 60 hari dari Bangkok ke Bali", desc: "Rute untuk pengembara digital dan backpacker berpengalaman. Anggaran ketat, coworking, dan koneksi lokal. Multi-mata uang aktif." }
  }
};

files.forEach(file => {
  const lang = file.split('.')[0];
  // map jp to ja or vice-versa, wait file is jp.json but translations keys:
  const key = lang === 'jp' ? 'jp' : lang;
  
  const filePath = path.join(localesDir, file);
  if (!fs.existsSync(filePath)) return;
  
  const data = JSON.parse(fs.readFileSync(filePath, 'utf8'));
  data.inspiration = translations[key];
  
  fs.writeFileSync(filePath, JSON.stringify(data, null, 2));
  console.log(`Updated ${file}`);
});
