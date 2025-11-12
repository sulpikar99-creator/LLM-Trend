# 🚀 Panduan Lengkap Penggunaan LLM Trend Trading Platform

## 📋 Apa itu LLM Trend?

LLM Trend adalah platform trading otomatis yang menggunakan AI (Large Language Model) untuk membuat keputusan trading pada cryptocurrency. Platform ini menggabungkan analisis teknikal dengan kecerdasan buatan untuk trading di Binance.

---

## 🎯 Fitur Utama yang Sudah Dibangun

### ✅ **Fitur Yang Sudah Lengkap:**

1. **Sistem Autentikasi**
   - Registrasi user dengan beta code
   - Login/logout dengan JWT token
   - Password encryption (bcrypt)

2. **Manajemen Trader AI**
   - Buat multiple trader bots
   - Konfigurasi per-trader (symbol, interval, leverage)
   - Start/stop trader secara individual
   - Monitoring status real-time

3. **Risk Management**
   - Max drawdown protection
   - Position size limits
   - Leverage limits
   - Daily loss limits
   - Stop loss & take profit otomatis

4. **AI Decision Engine**
   - Integrasi dengan DeepSeek/OpenAI/Anthropic
   - Custom strategy prompts
   - Context-aware trading decisions
   - Decision logging lengkap

5. **Real-time WebSocket** ⚡ (BARU!)
   - Live position updates
   - Live order updates
   - Real-time PnL tracking
   - Trade execution notifications
   - Balance updates

6. **Analytics Dashboard**
   - Performance metrics (win rate, profit factor, Sharpe ratio)
   - Drawdown analysis
   - Monte Carlo simulation
   - Correlation analysis
   - Performance attribution by symbol/strategy

7. **Database & Persistence**
   - SQLite database
   - Performance tracking
   - Trade history
   - Decision records
   - User configurations

---

## 🚀 Cara Menggunakan (Step-by-Step)

### **Langkah 1: Akses Web Interface**

Buka browser dan akses:
```
http://localhost:3000
```

### **Langkah 2: Buat Akun (Sudah Selesai ✓)**

Anda sudah berhasil registrasi! Skip langkah ini.

### **Langkah 3: Konfigurasi AI API Key** ⚠️ **PENTING!**

Tanpa API key, trader tidak bisa berjalan!

1. **Pergi ke halaman Settings** (klik "Settings" di navigation bar)

2. **Scroll ke section "AI Configuration"**

3. **Dapatkan API Key:**
   - **Untuk DeepSeek** (Recommended - murah):
     - Kunjungi: https://platform.deepseek.com
     - Buat akun
     - Generate API key
     - Copy API key (format: `sk-...`)

   - **Untuk OpenAI** (Lebih mahal):
     - Kunjungi: https://platform.openai.com
     - Buat akun
     - Generate API key

4. **Input konfigurasi:**
   - **AI Provider**: Pilih `deepseek` atau `openai`
   - **Model ID**: `deepseek-chat` (untuk DeepSeek) atau `gpt-4` (untuk OpenAI)
   - **API Key**: Paste API key Anda
   - **Base URL**: `https://api.deepseek.com` (untuk DeepSeek)
   - **Max Tokens**: `4000`
   - **Temperature**: `0.7`

5. **Klik "Save AI Configuration"**

### **Langkah 4: Konfigurasi Risk Limits** (Opsional tapi Disarankan)

Di halaman Settings, section "Risk Limits":

- **Max Drawdown**: `20%` (sistem akan stop jika loss mencapai 20%)
- **Max Position Size**: `$1000` (maksimal size per trade)
- **Max Leverage**: `10x` (maksimal leverage)
- **Daily Loss Limit**: `$500` (maksimal loss per hari)

Klik "Save Risk Limits"

### **Langkah 5: Buat Trader Pertama**

1. **Klik "Traders" di navigation bar**

2. **Klik tombol "Add Trader"** (atau tombol create trader)

3. **Isi form:**

   **Basic Information:**
   - **Name**: `Bitcoin Scalper` (atau nama apapun)
   - **Exchange**: Pilih `binance`
   - **Symbol**: `BTCUSDT` (atau symbol lain seperti `ETHUSDT`)
   - **Timeframe**: `15m` (atau `5m`, `1h`, `4h`, dll)

   **Exchange Configuration:**
   - **API Key**: API key Binance Anda
   - **API Secret**: Secret key Binance Anda
   - **Testnet**: ✅ **CENTANG INI** untuk testing dulu (sangat disarankan!)

   **⚠️ Cara Dapatkan Binance API Key:**
   - **Testnet (Untuk Testing - GRATIS)**:
     1. Kunjungi: https://testnet.binance.vision/
     2. Login dengan GitHub
     3. Generate API key di dashboard
     4. Dapat $100,000 virtual USDT

   - **Mainnet (Real Money - HATI-HATI)**:
     1. Kunjungi: https://www.binance.com
     2. Account → API Management
     3. Create API key
     4. Enable Spot & Futures Trading
     5. Whitelist IP jika perlu

   **Trading Parameters:**
   - **Initial Balance**: `10000` (balance awal)
   - **Max Positions**: `3` (maksimal 3 posisi bersamaan)
   - **Leverage**: `5` (leverage 5x)
   - **Position Size %**: `10` (gunakan 10% balance per trade)

   **Strategy Prompt:**
   ```
   You are a cryptocurrency trading expert. Analyze the market data and make trading decisions.

   Rules:
   - Only trade when there's high confidence
   - Use technical indicators (RSI, MACD, Moving Averages)
   - Consider market trends and volume
   - Risk/reward ratio minimum 1:2
   - Set stop loss 2% below entry
   - Set take profit 4% above entry

   Output format: JSON with action (buy/sell/hold), reason, entry_price, stop_loss, take_profit
   ```

4. **Klik "Create Trader"**

### **Langkah 6: Start Trading**

1. **Di halaman Traders**, Anda akan lihat trader yang baru dibuat

2. **Klik tombol "Start"** pada trader tersebut

3. **Trader akan mulai:**
   - Mengambil market data setiap interval (5m/15m/1h)
   - Menganalisis dengan AI
   - Membuat keputusan trading
   - Eksekusi order ke Binance
   - Update real-time via WebSocket

### **Langkah 7: Monitor Dashboard**

1. **Klik "Dashboard" di navigation bar**

2. **Anda akan melihat:**
   - Total balance
   - Active positions
   - Today's PnL
   - Recent trades
   - Performance chart

3. **Real-time Updates:**
   - Posisi update otomatis (WebSocket)
   - Order fills langsung muncul
   - PnL berubah real-time

### **Langkah 8: Lihat Analytics**

1. **Klik "Analytics" di navigation bar**

2. **4 Tab tersedia:**

   **Performance Tab:**
   - Win rate, profit factor, Sharpe ratio
   - Top performers by symbol
   - Average win/loss
   - Max streak

   **Drawdown Tab:**
   - Maximum drawdown
   - Current drawdown
   - Recovery time
   - Equity curve

   **Monte Carlo Tab:**
   - Probability of profit
   - Value at Risk (VaR)
   - Best/worst case scenarios
   - Confidence intervals

   ⚠️ **Note**: Perlu minimal 10-20 trades untuk data Monte Carlo

   **Correlation Tab:**
   - Correlation matrix antar symbols
   - Top correlations
   - Heatmap (jika ada multiple symbols)

---

## 🔍 Penjelasan Setiap Halaman

### 1. **Dashboard** (`/`)
**Fungsi:** Overview keseluruhan trading performance
- Total balance & equity
- Active positions list
- Today's profit/loss
- Recent trades history
- Quick stats

**Cara Pakai:**
- Lihat sekilas performa Anda
- Monitor posisi aktif
- Track PnL harian

### 2. **Traders** (`/traders`)
**Fungsi:** Manage multiple trading bots
- List semua traders
- Create new trader
- Start/stop trader
- Edit/delete trader
- Lihat status (running/stopped)

**Cara Pakai:**
- Buat trader baru untuk symbol berbeda
- Start/stop sesuai kebutuhan
- Monitor status masing-masing trader

### 3. **Analytics** (`/analytics`)
**Fungsi:** Deep dive analysis trading performance
- **Performance**: Metrics lengkap
- **Drawdown**: Risk analysis
- **Monte Carlo**: Probability simulation
- **Correlation**: Inter-symbol correlation

**Cara Pakai:**
- Analisis performa setelah punya trading history
- Identifikasi symbol terbaik/terburuk
- Evaluate risk dengan drawdown analysis
- Simulate future scenarios dengan Monte Carlo

### 4. **Settings** (`/settings`)
**Fungsi:** Konfigurasi platform
- **Risk Limits**: Max drawdown, position size, leverage, daily loss
- **AI Configuration**: Provider, model, API key
- **Strategy Prompt**: Custom trading strategy

**Cara Pakai:**
- Setup AI API key (WAJIB!)
- Set risk limits untuk protect capital
- Customize strategy prompt

---

## 🎨 Kenapa Halaman Terlihat Kosong?

Jika Anda baru registrasi dan belum buat trader, halaman akan menampilkan **"empty state"**:

### Dashboard:
```
"No active traders. Create a trader to start trading!"
```

### Traders Page:
```
"No traders yet. Click 'Add Trader' to create your first trading bot."
```

### Analytics Page:
```
"No trading data available yet. Start trading to see analytics!"
```

**Ini NORMAL!** Anda perlu:
1. ✅ Konfigurasi AI API key di Settings
2. ✅ Buat trader di Traders page
3. ✅ Start trader
4. ⏳ Tunggu beberapa trades tereksekusi
5. 🎉 Data akan muncul!

---

## 🔧 Troubleshooting

### ❌ "API Key Required" di Settings
**Solusi:** Input API key dari DeepSeek atau OpenAI di Settings → AI Configuration

### ❌ Trader tidak start
**Penyebab:**
- API key belum diset
- Binance API credentials salah
- Balance tidak cukup

**Solusi:**
- Cek Settings → AI Configuration sudah diisi
- Verify Binance API key & secret
- Gunakan testnet untuk testing

### ❌ "No historical trading data" di Monte Carlo
**Ini Normal!** Monte Carlo perlu minimal 10-20 trades untuk running simulation.

**Solusi:** Tunggu trader jalan dan eksekusi beberapa trades dulu.

### ❌ Dashboard kosong
**Penyebab:** Belum ada trader yang running

**Solusi:**
1. Buat trader di Traders page
2. Start trader
3. Tunggu beberapa trading cycles

### ❌ WebSocket disconnected
**Solusi:** Refresh halaman. WebSocket akan auto-reconnect.

---

## 📊 Apa yang Bisa Dilakukan Sekarang?

### ✅ **Sudah Berfungsi:**
1. User authentication (login/register)
2. AI configuration (DeepSeek/OpenAI/Anthropic)
3. Risk limits configuration
4. Create/manage multiple traders
5. Start/stop traders
6. Real-time WebSocket updates
7. Performance tracking
8. Analytics & statistics
9. Drawdown analysis
10. Monte Carlo simulation
11. Correlation analysis
12. Decision logging

### 🔄 **Cara Kerja Sistem:**

```
1. User set AI API key di Settings
   ↓
2. User buat Trader dengan symbol (e.g., BTCUSDT)
   ↓
3. User klik Start pada trader
   ↓
4. Trader mulai cycle:
   - Fetch market data (klines, orderbook, trades)
   - Kirim data ke AI untuk analisis
   - AI return decision (buy/sell/hold)
   - Eksekusi order ke Binance
   - Update position & PnL
   - Broadcast via WebSocket
   - Save ke database
   ↓
5. Dashboard update real-time
   ↓
6. Analytics populate dengan data
```

---

## 💡 Tips Penggunaan

### 1. **Mulai dengan Testnet**
- Gunakan Binance Testnet dulu
- Gratis $100k virtual USDT
- Zero risk untuk testing

### 2. **Set Risk Limits Ketat**
- Max drawdown: 10-20%
- Position size: 5-10% of balance
- Leverage: 3-5x (jangan terlalu tinggi)

### 3. **Monitor Awal**
- Awasi 10-20 trades pertama
- Pastikan AI decision make sense
- Adjust strategy prompt jika perlu

### 4. **Diversifikasi**
- Buat multiple traders
- Different symbols (BTC, ETH, BNB)
- Different timeframes (5m, 15m, 1h)

### 5. **Review Analytics**
- Cek performance metrics mingguan
- Identifikasi symbol terbaik
- Adjust strategy based on data

---

## 🚨 Peringatan Penting

1. **Trading cryptocurrency berisiko tinggi** - Hanya gunakan uang yang siap Anda hilangkan
2. **Gunakan Testnet dulu** - Jangan langsung mainnet
3. **Set stop loss** - Selalu gunakan risk management
4. **Monitor reguler** - Jangan tinggalkan bot tanpa monitoring
5. **API key security** - Jangan share API key Anda, database dienkripsi AES-256

---

## 📞 Support

Jika ada pertanyaan atau masalah:
1. Cek log dengan: `docker-compose logs -f`
2. Cek browser console (F12) untuk JavaScript errors
3. Verify API keys sudah benar
4. Pastikan Binance API key punya permission trading

---

## 🎯 Checklist Quick Start

- [ ] Akses http://localhost:3000
- [ ] Login dengan akun yang sudah dibuat
- [ ] Pergi ke Settings
- [ ] Input DeepSeek/OpenAI API key
- [ ] Set risk limits
- [ ] Pergi ke Traders page
- [ ] Klik "Add Trader"
- [ ] Isi form dengan Binance Testnet credentials
- [ ] Klik "Create Trader"
- [ ] Klik "Start" pada trader
- [ ] Pergi ke Dashboard untuk monitor
- [ ] Tunggu beberapa trades
- [ ] Cek Analytics untuk metrics

---

**Selamat Trading! 🚀📈**

Platform sudah **LENGKAP dan SIAP DIGUNAKAN**. Silakan ikuti langkah-langkah di atas untuk mulai trading dengan AI.
