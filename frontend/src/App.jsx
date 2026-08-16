import { useState, useEffect } from 'react'
import './index.css'

function App() {
  const [token, setToken] = useState(localStorage.getItem('token'))
  const [userName, setUserName] = useState(localStorage.getItem('name'))
  const [email, setEmail] = useState('dinesh@test.com')
  const [password, setPassword] = useState('password123')
  
  const [turfs, setTurfs] = useState([])
  const [slots, setSlots] = useState([])
  const [userBookings, setUserBookings] = useState([])
  
  // views: 'home' | 'explore' | 'bookings' | 'profile' | 'overview' | 'slots' | 'checkout' | 'success' | 'booking-details'
  const [view, setView] = useState('home') 
  const [selectedTurf, setSelectedTurf] = useState(null)
  const [selectedSlot, setSelectedSlot] = useState(null)
  const [selectedBooking, setSelectedBooking] = useState(null)
  
  const [searchQuery, setSearchQuery] = useState('')
  const [activeCategory, setActiveCategory] = useState('All')
  const [payAdvance, setPayAdvance] = useState(false)
  const [selectedDate, setSelectedDate] = useState(new Date().toISOString().split('T')[0])
  const [splitSessionId, setSplitSessionId] = useState(null)
  
  const [timeLeft, setTimeLeft] = useState({ days: '00', hrs: '00', min: '00', sec: '00' })

  const [toast, setToast] = useState(null)
  const [loading, setLoading] = useState(false)
  const [bookingLoading, setBookingLoading] = useState(false)

  const fetchData = () => {
    if (token) {
      fetch('/api/turfs', { headers: { 'Authorization': `Bearer ${token}` } })
        .then(r => r.json()).then(setTurfs).catch(() => {})
      fetch('/api/slots', { headers: { 'Authorization': `Bearer ${token}` } })
        .then(r => r.json()).then(setSlots).catch(() => {})
      fetch('/api/user/bookings', { headers: { 'Authorization': `Bearer ${token}` } })
        .then(r => r.json()).then(setUserBookings).catch(() => {})
    }
  }

  useEffect(() => {
    fetchData()
  }, [token])

  useEffect(() => {
    const now = new Date()
    const upcomingBookings = userBookings.filter(b => new Date(b.start_time) > now && b.status !== 'CANCELLED')
    upcomingBookings.sort((a, b) => new Date(a.start_time) - new Date(b.start_time))
    const nextBooking = upcomingBookings.length > 0 ? upcomingBookings[0] : null

    if (!nextBooking) return

    const interval = setInterval(() => {
      const currentTime = new Date()
      const startTime = new Date(nextBooking.start_time)
      const diff = startTime - currentTime

      if (diff <= 0) {
        clearInterval(interval)
        setTimeLeft({ days: '00', hrs: '00', min: '00', sec: '00' })
        return
      }
      
      const days = Math.floor(diff / (1000 * 60 * 60 * 24))
      const hrs = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60))
      const min = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))
      const sec = Math.floor((diff % (1000 * 60)) / 1000)
      
      setTimeLeft({
        days: days.toString().padStart(2, '0'),
        hrs: hrs.toString().padStart(2, '0'),
        min: min.toString().padStart(2, '0'),
        sec: sec.toString().padStart(2, '0')
      })
    }, 1000)

    return () => clearInterval(interval)
  }, [userBookings])

  const showToast = (type, message) => {
    setToast({ type, message })
    setTimeout(() => setToast(null), 4000)
  }

  const handleLogin = async (e) => {
    e.preventDefault()
    setLoading(true)
    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ email, password })
      })
      const data = await response.json()
      if (response.ok) {
        setToken(data.token)
        setUserName(data.name)
        localStorage.setItem('token', data.token)
        localStorage.setItem('name', data.name)
      } else { showToast('error', data || 'Login failed') }
    } catch (err) { showToast('error', 'Server error') } finally { setLoading(false) }
  }

  const handleLogout = () => {
    setToken(null); setUserName(null); localStorage.clear(); setView('home')
  }

  const handleBook = async () => {
    if (!selectedSlot) return
    setBookingLoading(true)
    const response = await fetch('/api/bookings', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
      body: JSON.stringify({ slot_id: selectedSlot.id }) 
    })
    setBookingLoading(false)
    if (response.ok) {
      fetchData()
      setView('success')
    } else {
      const data = await response.json()
      showToast('error', data.message || data.error)
    }
  }

  const handleTeamBooking = async () => {
    if (!selectedSlot || !selectedTurf) return
    setBookingLoading(true)
    try {
      const response = await fetch('/api/team-sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` },
        body: JSON.stringify({ slot_id: selectedSlot.id, target_amount: selectedTurf.price_per_hour }) 
      })
      setBookingLoading(false)
      const data = await response.json()
      if (response.ok) {
        showToast('success', 'Team session created! ID: ' + data.session_id)
        setSplitSessionId(data.session_id)
        setView('split-share')
      } else {
        showToast('error', data.message || data.error)
      }
    } catch (err) {
      setBookingLoading(false)
      showToast('error', 'Network error.')
    }
  }

  const handleCancelBooking = async () => {
    if (!selectedBooking) return
    setBookingLoading(true)
    try {
      const response = await fetch(`/api/bookings/${selectedBooking.booking_id}/cancel`, {
        method: 'POST', headers: { 'Authorization': `Bearer ${token}` }
      })
      setBookingLoading(false)
      if (response.ok) {
        showToast('success', 'Booking cancelled.')
        fetchData(); setView('bookings')
      } else {
        showToast('error', 'Failed to process cancellation. Please try again.')
      }
    } catch (err) {
      setBookingLoading(false)
      showToast('error', 'Failed to process cancellation. Please try again.')
    }
  }

  const getTurfImage = (index) => index % 2 === 0 ? '/assets/indoor_turf.jpg' : '/assets/outdoor_turf.jpg'
  const getSportType = (index) => index % 2 === 0 ? 'Football' : 'Cricket'

  // --- COMPONENTS ---
  const BottomNav = () => (
    <div className="bottom-nav nav-bar">
      <div className={`nav-item ${view === 'home' ? 'active' : ''}`} onClick={() => setView('home')}>
        <div style={{ fontSize: '20px' }}>🏠</div> Home
      </div>
      <div className={`nav-item ${view === 'explore' ? 'active' : ''}`} onClick={() => setView('explore')}>
        <div style={{ fontSize: '20px' }}>🔍</div> Explore
      </div>
      <div className={`nav-item ${view === 'bookings' ? 'active' : ''}`} onClick={() => setView('bookings')}>
        <div style={{ fontSize: '20px' }}>🔖</div> Bookings
      </div>
      <div className={`nav-item ${view === 'profile' ? 'active' : ''}`} onClick={() => setView('profile')}>
        <div style={{ fontSize: '20px' }}>👤</div> Profile
      </div>
    </div>
  )

  if (!token) {
    return (
      <div className="login-container">
        <div className="login-card">
          <h2 style={{ textAlign: 'center', margin: '0 0 30px 0', fontSize: '28px' }}>TURF<span style={{ color: 'var(--primary-color)' }}>BAY</span></h2>
          <form onSubmit={handleLogin}>
            <input type="email" placeholder="Email" className="input-field" value={email} onChange={e => setEmail(e.target.value)} required />
            <input type="password" placeholder="Password" className="input-field" value={password} onChange={e => setPassword(e.target.value)} required />
            <button type="submit" className="fbb-btn" style={{ width: '100%' }} disabled={loading}>{loading ? 'Signing in...' : 'Login'}</button>
          </form>
          {toast && <div style={{ color: '#ff5a5f', textAlign: 'center', marginTop: '15px' }}>{toast.message}</div>}
        </div>
      </div>
    )
  }

  // --- VIEWS ---
  const renderHome = () => {
    const now = new Date()
    const upcomingBookings = userBookings.filter(b => new Date(b.start_time) > now && b.status !== 'CANCELLED')
    upcomingBookings.sort((a, b) => new Date(a.start_time) - new Date(b.start_time))
    const upcoming = upcomingBookings.length > 0 ? upcomingBookings[0] : null

    return (
      <>
        <div className="top-header" style={{ marginBottom: '20px' }}>
          <div className="logo">TURF<span style={{ color: 'var(--primary-color)' }}>BAY</span></div>
          <div className="bell-icon">🔔</div>
        </div>

        <div className="welcome-pill">
          Welcome back, {userName ? userName.split(' ')[0] : 'User'}! 👋
        </div>

        <h1 className="home-epic-title">
          Let's play<br/>
          <span>something epic</span><br/>
          today.
        </h1>

        <div className="home-search-bar" onClick={() => setView('explore')}>
          <div style={{ color: 'var(--text-secondary)' }}>📍</div>
          <input type="text" placeholder="Search location or turf..." readOnly />
          <button className="home-search-btn">🔍</button>
        </div>

        <div className="section-header">
          <h2 className="section-title">Upcoming Booking</h2>
        </div>

        {upcoming ? (
          <div className="upcoming-full-card" onClick={() => { setSelectedBooking(upcoming); setView('booking-details') }}>
            <div className="upcoming-img-section">
              <img src="/assets/indoor_turf.jpg" alt="turf" onError={(e) => { e.target.onerror = null; e.target.src = "https://images.unsplash.com/photo-1518605368461-1ee7e5436660?q=80&w=300&auto=format&fit=crop"; }} />
              <div className="upcoming-date-badge">
                <div className="month">{new Date(upcoming.start_time).toLocaleString('default', { month: 'short' })}</div>
                <div className="day">{new Date(upcoming.start_time).getDate()}</div>
              </div>
            </div>
            
            <div className="upcoming-details-section">
              <div className="next-booking-badge">Next Booking</div>
              <h3>{upcoming.turf_name}</h3>
              <p>📍 {upcoming.location || 'Unknown Location'}</p>
              <p>
                {new Date(upcoming.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })} - {new Date(upcoming.end_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </p>
              
              <p style={{ marginTop: '10px' }}>Starts In</p>
              <div className="countdown-container">
                {timeLeft.days !== '00' && (
                  <div className="countdown-item">
                    <span className="countdown-val">{timeLeft.days}</span>
                    <span className="countdown-label">DAYS</span>
                  </div>
                )}
                <div className="countdown-item">
                  <span className="countdown-val">{timeLeft.hrs}</span>
                  <span className="countdown-label">HRS</span>
                </div>
                <div className="countdown-item">
                  <span className="countdown-val">{timeLeft.min}</span>
                  <span className="countdown-label">MIN</span>
                </div>
                <div className="countdown-item">
                  <span className="countdown-val">{timeLeft.sec}</span>
                  <span className="countdown-label">SEC</span>
                </div>
              </div>
            </div>
          </div>
        ) : (
          <div className="upcoming-card">
            <div className="upcoming-icon">📅</div>
            <h3 style={{ margin: 0, fontSize: '16px' }}>No upcoming games</h3>
            <p style={{ margin: 0, fontSize: '12px', color: 'var(--text-secondary)' }}>Book a turf to see your schedule here.</p>
          </div>
        )}

        <div className="section-header">
          <h2 className="section-title">Popular Turfs</h2>
          <span className="see-all" onClick={() => setView('explore')}>See All</span>
        </div>
        <div className="turf-scroll-row">
          {turfs.slice(0,4).map((turf, i) => (
            <div key={turf.id} className="turf-card-home turf-card" onClick={() => { setSelectedTurf({...turf, imageIndex: i}); setView('overview') }}>
              <div className="badge-overlay">{getSportType(i)}</div>
              <img src={getTurfImage(i)} alt={turf.name} className="turf-image" />
              <div className="turf-info">
                <h3>{turf.name}</h3>
                <p>📍 {turf.location}</p>
                <div className="turf-price-row">
                  <div>
                    <div className="price-label">Starting from</div>
                    <div className="price-val">₹{turf.price_per_hour}/hour</div>
                  </div>
                  <button className="book-btn-sm">Book</button>
                </div>
              </div>
            </div>
          ))}
        </div>
        <BottomNav />
      </>
    )
  }

  const renderExplore = () => {
    let filtered = turfs.filter(t => t.name.toLowerCase().includes(searchQuery.toLowerCase()) || t.location.toLowerCase().includes(searchQuery.toLowerCase()))
    return (
      <>
        <div style={{ padding: '20px' }}>
          <div className="search-container" style={{ margin: 0 }}>
            <span style={{ padding: '0 10px', color: '#94a3b8' }}>🔍</span>
            <input type="text" className="search-input" placeholder="Search turf, sport, location" value={searchQuery} onChange={e => setSearchQuery(e.target.value)} autoFocus />
            <button className="search-btn">→</button>
          </div>
        </div>
        
        <div className="category-pills">
          {['All', 'Football', 'Cricket', 'Basketball', 'Tennis'].map(cat => (
            <div key={cat} className={`cat-pill ${activeCategory === cat ? 'active' : ''}`} onClick={() => setActiveCategory(cat)}>
              {cat === 'All' ? '🎛 ' : ''}{cat}
            </div>
          ))}
        </div>

        <div className="section-header">
          <h2 className="section-title">Available Turfs</h2>
          <span style={{ background: 'rgba(163,230,53,0.1)', color: 'var(--primary-color)', padding: '4px 10px', borderRadius: '12px', fontSize: '12px' }}>{filtered.length} found</span>
        </div>

        <div className="turf-list">
          {filtered.map((turf, i) => (
            <div key={turf.id} className="turf-list-card" onClick={() => { setSelectedTurf({...turf, imageIndex: i}); setView('overview') }}>
              <img src={getTurfImage(i)} alt={turf.name} className="turf-list-img" />
              <div className="turf-list-info">
                <div>
                  <h4>{turf.name}</h4>
                  <p style={{ margin: '5px 0 0 0', fontSize: '12px', color: 'var(--text-secondary)' }}>📍 {turf.location} • {getSportType(i)}</p>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
                  <div>
                    <div style={{ fontSize: '10px', color: 'var(--text-secondary)' }}>Price / hr</div>
                    <div style={{ fontSize: '14px', fontWeight: 700, color: 'var(--primary-color)' }}>₹{turf.price_per_hour}/hour</div>
                  </div>
                  <button className="book-btn-sm">Book</button>
                </div>
              </div>
            </div>
          ))}
        </div>
        <BottomNav />
      </>
    )
  }

  const renderOverview = () => {
    if (!selectedTurf) return null
    return (
      <>
        <div className="overview-img-container">
          <img src={getTurfImage(selectedTurf.imageIndex)} alt={selectedTurf.name} className="overview-img-main" />
          <div className="top-actions">
            <button className="icon-btn" onClick={() => setView('home')}>←</button>
            <button className="icon-btn">↗</button>
          </div>
        </div>
        
        <div className="overview-content">
          <div className="pill-row">
            <span className="pill" style={{ borderColor: 'var(--primary-color)', color: 'white' }}>🏟 {getSportType(selectedTurf.imageIndex)}</span>
          </div>
          <h1 style={{ margin: '0 0 20px 0', fontSize: '28px' }}>{selectedTurf.name}</h1>
          
          <div className="pill-row">
            <span className="pill">Cricket</span>
            <span className="pill">Floodlights</span>
            <span className="pill">Parking</span>
            <span className="pill">Online slots</span>
          </div>

          <div className="info-boxes">
            <div className="info-box">
              <div className="info-box-title">⭐ Google rating</div>
              <div className="info-box-val">No ratings</div>
            </div>
            <div className="info-box">
              <div className="info-box-title">₹ Price</div>
              <div className="info-box-val">₹{selectedTurf.price_per_hour}/hour</div>
            </div>
          </div>

          <div className="offer-banner">
            🏷 10% Off at weekend
          </div>

          <h3 style={{ margin: '0 0 10px 0' }}>About this turf</h3>
          <p style={{ margin: '0 0 20px 0', color: 'var(--text-secondary)', lineHeight: 1.5, fontSize: '14px' }}>
            Experience the best-in-class sports infrastructure at {selectedTurf.name}. 
            Perfect for professional matches and casual games with friends.
          </p>

          <h3 style={{ margin: '0 0 10px 0' }}>Location</h3>
          <div style={{ background: 'var(--card-bg)', border: '1px solid var(--border-color)', padding: '15px', borderRadius: '12px', display: 'flex', alignItems: 'center', gap: '10px' }}>
            <span style={{ color: 'var(--primary-color)' }}>◎</span> Location Details <br/>
            <span style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>{selectedTurf.location}</span>
          </div>
        </div>

        <div className="floating-bottom-bar">
          <div>
            <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Starting from</div>
            <div style={{ fontSize: '20px', fontWeight: 700, color: 'var(--primary-color)' }}>₹{selectedTurf.price_per_hour}/hour</div>
          </div>
          <button className="fbb-btn" onClick={() => setView('slots')}>📅 BOOK NOW</button>
        </div>
      </>
    )
  }

  const renderSlots = () => {
    if (!selectedTurf) return null
    
    // Generate next 7 days for the date selector
    const dateOptions = Array.from({length: 7}).map((_, i) => {
      const d = new Date()
      d.setDate(d.getDate() + i)
      return {
        fullDate: d.toISOString().split('T')[0],
        dayName: i === 0 ? 'TODAY' : d.toLocaleDateString('en-US', { weekday: 'short' }).toUpperCase(),
        dayNum: d.getDate().toString().padStart(2, '0')
      }
    })

    // Filter slots for the selected turf AND the selected date
    const turfSlots = slots.filter(s => {
      if (s.turf_id !== selectedTurf.id) return false
      const slotDate = s.start_time.split('T')[0]
      return slotDate === selectedDate
    })

    return (
      <>
        <div className="top-header" style={{ padding: '20px 20px 10px 20px' }}>
          <button className="icon-btn" style={{ background: 'transparent' }} onClick={() => setView('overview')}>←</button>
          <div style={{ flex: 1, marginLeft: '15px' }}>
            <h2 style={{ margin: 0, fontSize: '18px' }}>{selectedTurf.name}</h2>
            <p style={{ margin: 0, fontSize: '12px', color: 'var(--text-secondary)' }}>{selectedTurf.location}</p>
          </div>
        </div>
        
        <div style={{ padding: '0 20px 10px 20px' }}>
          <div style={{ color: '#facc15', fontSize: '12px', fontWeight: 600, display: 'flex', alignItems: 'center', gap: '5px' }}>⭐ New <span style={{ color: 'var(--text-secondary)', fontWeight: 'normal' }}>On day time 20% offer</span></div>
        </div>

        <div style={{ padding: '10px 20px' }}>
          <h3 style={{ margin: '0 0 5px 0', fontSize: '16px' }}>Select date</h3>
          <p style={{ margin: 0, fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '15px' }}>Choose your preferred playing day</p>
        </div>
        
        <div className="date-selector">
          {dateOptions.map(opt => (
            <div key={opt.fullDate} 
                 className={`date-box ${selectedDate === opt.fullDate ? 'active' : ''}`} 
                 onClick={() => { setSelectedDate(opt.fullDate); setSelectedSlot(null); }}>
              <div className="date-day">{opt.dayName}</div>
              <div className="date-num">{opt.dayNum}</div>
            </div>
          ))}
        </div>

        <div style={{ padding: '10px 20px', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <div>
            <h3 style={{ margin: '0 0 5px 0', fontSize: '16px' }}>Available slots</h3>
            <p style={{ margin: 0, fontSize: '12px', color: 'var(--text-secondary)' }}>Tap one or more available slots</p>
          </div>
          <div style={{ display: 'flex', gap: '10px', fontSize: '12px', color: 'var(--text-secondary)' }}>
            <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}><div style={{ width: '8px', height: '8px', borderRadius: '50%', background: 'var(--primary-color)' }}></div> Available</span>
            <span style={{ display: 'flex', alignItems: 'center', gap: '4px' }}><div style={{ width: '8px', height: '8px', borderRadius: '50%', background: '#475569' }}></div> Unavailable</span>
          </div>
        </div>

        <div className="slots-grid">
          {turfSlots.map(slot => {
            const timeStr = new Date(slot.start_time).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
            const isSelected = selectedSlot?.id === slot.id
            return (
              <div key={slot.id} className={`slot slot-item ${isSelected ? 'active' : ''} ${slot.is_booked ? 'disabled booked' : ''}`}
                onClick={() => { if (!slot.is_booked) { setSelectedSlot(slot) } }}
              >
                {timeStr}
              </div>
            )
          })}
        </div>

        <div className="floating-bottom-bar">
          <div>
            <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Total Estimate</div>
            <div style={{ fontSize: '20px', fontWeight: 700, color: 'var(--primary-color)' }}>₹{selectedSlot ? selectedTurf.price_per_hour : 0}</div>
          </div>
          <button className="fbb-btn" onClick={() => selectedSlot && setView('checkout')} style={{ opacity: selectedSlot ? 1 : 0.5 }}>
            PROCEED
          </button>
        </div>
      </>
    )
  }

  const renderCheckout = () => {
    if (!selectedTurf || !selectedSlot) return null
    const dateObj = new Date(selectedSlot.start_time)
    const dateStr = dateObj.toLocaleDateString([], { weekday: 'short', day: '2-digit', month: 'short', year: 'numeric' }).toUpperCase()
    const timeStr = dateObj.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    const slotPrice = selectedTurf.price_per_hour
    const fee = 45; const total = slotPrice + fee; const advanceAmount = Math.ceil(total / 2)

    return (
      <>
        <div className="top-header" style={{ justifyContent: 'center', position: 'relative' }}>
          <button className="icon-btn" style={{ position: 'absolute', left: '20px', background: 'transparent' }} onClick={() => setView('slots')}>←</button>
          <h2 style={{ margin: 0, fontSize: '18px' }}>Confirm Booking</h2>
        </div>

        <div className="checkout-details">
          <div className="checkout-card">
            <img src={getTurfImage(selectedTurf.imageIndex)} alt={selectedTurf.name} className="checkout-img" />
            <div>
              <h3 style={{ margin: '0 0 5px 0', fontSize: '16px' }}>{selectedTurf.name}</h3>
              <p style={{ margin: 0, fontSize: '12px', color: 'var(--text-secondary)' }}>{selectedTurf.location}</p>
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '15px', marginBottom: '30px', color: 'var(--text-secondary)', fontSize: '14px' }}>
            <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}><span>📅</span> {dateStr}</div>
            <div style={{ display: 'flex', gap: '10px', alignItems: 'center' }}><span>🕒</span> {timeStr}</div>
          </div>

          <h3 style={{ margin: '0 0 15px 0', fontSize: '16px' }}>Estimated Billing</h3>
          <div className="bill-row"><span>Slot Price</span><span>₹{slotPrice}</span></div>
          <div className="bill-row"><span>Convenience Fee</span><span>₹{fee}</span></div>
          <div className="bill-row total"><span>Grand Total</span><span>₹{total}</span></div>

          <div style={{ marginTop: '30px' }} className={`advance-box ${payAdvance ? 'active' : ''}`} onClick={() => setPayAdvance(!payAdvance)}>
            <div style={{ width: '24px', height: '24px', borderRadius: '50%', border: '2px solid', borderColor: payAdvance ? 'var(--primary-color)' : 'var(--border-color)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              {payAdvance && <div style={{ width: '12px', height: '12px', borderRadius: '50%', background: 'var(--primary-color)' }}></div>}
            </div>
            <div>
              <div style={{ fontSize: '14px', fontWeight: 600, color: payAdvance ? 'var(--primary-color)' : 'white' }}>Pay Advance Only</div>
              <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>Pay ₹{advanceAmount} now, pay ₹{total - advanceAmount} at the turf.</div>
            </div>
          </div>
        </div>

        <div className="floating-bottom-bar">
          <div>
            <div style={{ fontSize: '12px', color: 'var(--text-secondary)' }}>To Pay Now</div>
            <div style={{ fontSize: '20px', fontWeight: 700, color: 'var(--primary-color)' }}>₹{payAdvance ? advanceAmount : total}</div>
          </div>
          <div style={{ display: 'flex', gap: '10px' }}>
            <button className="fbb-btn" style={{ background: '#334155' }} onClick={handleTeamBooking} disabled={bookingLoading}>
              {bookingLoading ? '...' : 'TEAM SPLIT'}
            </button>
            <button className="fbb-btn" onClick={handleBook} disabled={bookingLoading}>
              {bookingLoading ? '...' : 'PROCEED'}
            </button>
          </div>
        </div>
      </>
    )
  }

  const renderSplitShare = () => {
    const shareLink = `${window.location.origin}/split-pay/${splitSessionId}`
    return (
      <div style={{ padding: '20px', textAlign: 'center', marginTop: '50px' }}>
        <h2 style={{ color: 'var(--primary-color)' }}>Team Session Created!</h2>
        <p style={{ color: 'var(--text-secondary)' }}>Share this unique link with your friends to collect payments.</p>
        <p style={{ color: 'var(--text-secondary)', fontSize: '12px' }}>The slot is held in PENDING state for 15 minutes.</p>
        
        <div style={{ background: '#1e293b', padding: '15px', borderRadius: '10px', marginTop: '20px', wordBreak: 'break-all', fontFamily: 'monospace', color: '#fff' }}>
          {shareLink}
        </div>
        
        <button className="primary-btn" style={{ marginTop: '30px' }} onClick={() => { setView('bookings'); fetchData(); }}>
          VIEW MY BOOKINGS
        </button>
      </div>
    )
  }

  const renderBookings = () => (
    <>
      <div className="page-header">My Bookings</div>
      <div className="booking-list">
        {userBookings.length === 0 ? (
          <div style={{ textAlign: 'center', color: 'var(--text-secondary)' }}>No bookings found.</div>
        ) : (
          userBookings.map((b, i) => {
            const isCancelled = b.status === 'CANCELLED'
            const isPartial = i % 2 !== 0
            const dateStr = new Date(b.start_time).toLocaleDateString()
            return (
              <div key={b.booking_id} className="bk-card" onClick={() => { setSelectedBooking(b); setView('booking-details') }}>
                <div className="bk-header">
                  <img src={getTurfImage(i)} alt="turf" className="bk-img" />
                  <div className="bk-title">
                    <h4>{b.turf_name}</h4>
                    <p>📍 {b.location}</p>
                  </div>
                  <div className={`bk-status ${b.status === 'CANCELLED' ? 'cancelled' : b.status === 'PENDING' ? 'partial' : 'full'}`}>
                    {b.status === 'PENDING' ? 'PENDING' : b.status === 'CANCELLED' ? 'CANCELLED' : 'FULL PAID'}
                  </div>
                </div>
                <div className="bk-body">
                  <div className="bk-info-block">
                    <div className="bk-icon">📅</div>
                    <div className="bk-val"><p>Game Date</p><h5>{dateStr}</h5></div>
                  </div>
                  <div className="bk-info-block">
                    <div className="bk-icon">₹</div>
                    <div className="bk-val"><p>Amount Paid</p><h5>₹{isPartial ? '772' : '1545'}</h5></div>
                  </div>
                </div>
              </div>
            )
          })
        )}
      </div>
      <BottomNav />
    </>
  )

  const renderBookingDetails = () => {
    if (!selectedBooking) return null
    const b = selectedBooking
    const dateStr = new Date(b.start_time).toLocaleDateString()
    const timeStr = `${new Date(b.start_time).toLocaleTimeString([], {hour:'2-digit', minute:'2-digit', hour12:false})} - ${new Date(b.end_time).toLocaleTimeString([], {hour:'2-digit', minute:'2-digit', hour12:false})}`

    return (
      <>
        <div className="top-header" style={{ justifyContent: 'center', position: 'relative' }}>
          <button className="icon-btn" style={{ position: 'absolute', left: '20px', background: 'transparent' }} onClick={() => setView('bookings')}>←</button>
          <h2 style={{ margin: 0, fontSize: '18px' }}>My Bookings</h2>
        </div>
        
        <div className="sheet-container">
          <div className="sheet-card">
            <div className="bk-header" style={{ padding: '0 0 20px 0', borderBottom: 'none' }}>
              <img src={getTurfImage(0)} alt="turf" className="bk-img" />
              <div className="bk-title">
                <h4>{b.turf_name}</h4>
                <p>{b.location}</p>
              </div>
              <div className={`bk-status ${b.status === 'CANCELLED' ? 'cancelled' : 'full'}`}>{b.status}</div>
            </div>

            <div className="bk-body" style={{ padding: '15px 0' }}>
              <div className="bk-info-block">
                <div className="bk-icon">📅</div>
                <div className="bk-val"><p>Game Date</p><h5>{dateStr}</h5></div>
              </div>
              <div className="bk-info-block">
                <div className="bk-icon">₹</div>
                <div className="bk-val"><p>Paid (Full)</p><h5>₹1545</h5></div>
              </div>
            </div>

            <div style={{ marginTop: '10px' }}>
              <p style={{ fontSize: '12px', color: 'var(--text-secondary)', marginBottom: '10px' }}>Booked Slots</p>
              <div className="slot-pill">{timeStr}</div>
            </div>

            <div className="ref-box">
              <span>🎫</span> Ref: {b.booking_id}
            </div>
          </div>

          {b.status === 'CONFIRMED' && (
            <button className="cancel-btn" onClick={handleCancelBooking} disabled={bookingLoading}>
              ⊗ {bookingLoading ? 'Requesting...' : 'Request Cancellation'}
            </button>
          )}
        </div>
      </>
    )
  }

  const renderProfile = () => (
    <>
      <div className="page-header">Profile</div>
      <div style={{ padding: '20px', textAlign: 'center' }}>
        <div style={{ width: '100px', height: '100px', borderRadius: '50%', background: 'var(--card-bg)', border: '2px solid var(--primary-color)', margin: '0 auto 20px auto', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: '40px' }}>👤</div>
        <h2>{userName}</h2>
        <p style={{ color: 'var(--text-secondary)' }}>{email}</p>
        <button onClick={handleLogout} style={{ marginTop: '30px', background: 'transparent', border: '1px solid #ef4444', color: '#ef4444', padding: '10px 30px', borderRadius: '8px', cursor: 'pointer' }}>Sign Out</button>
      </div>
      <BottomNav />
    </>
  )

  const renderSuccess = () => (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh', padding: '20px', textAlign: 'center' }}>
      <div style={{ width: '80px', height: '80px', borderRadius: '50%', background: 'var(--primary-color)', display: 'flex', alignItems: 'center', justifyContent: 'center', color: 'var(--bg-color)', fontSize: '40px', marginBottom: '20px' }}>✓</div>
      <h2 style={{ color: 'var(--primary-color)' }}>Booking Confirmed!</h2>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '40px' }}>Your slot is reserved.</p>
      <button className="fbb-btn" style={{ width: '100%', marginBottom: '15px' }} onClick={() => setView('bookings')}>View Bookings</button>
      <button className="fbb-btn" style={{ width: '100%', background: 'transparent', border: '1px solid var(--primary-color)', color: 'var(--primary-color)' }} onClick={() => setView('home')}>Back to Home</button>
    </div>
  )

  return (
    <div>
      {view === 'home' && renderHome()}
      {view === 'explore' && renderExplore()}
      {view === 'overview' && renderOverview()}
      {view === 'slots' && renderSlots()}
      {view === 'checkout' && renderCheckout()}
      {view === 'split-share' && renderSplitShare()}
      {view === 'success' && renderSuccess()}
      {view === 'bookings' && renderBookings()}
      {view === 'booking-details' && renderBookingDetails()}
      {view === 'profile' && renderProfile()}
      {toast && <div className={`toast ${toast.type}`}>{toast.message}</div>}
    </div>
  )
}

export default App
