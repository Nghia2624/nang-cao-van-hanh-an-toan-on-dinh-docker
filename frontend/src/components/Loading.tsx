
export default function Loading({ message = 'Loading...' }: { message?: string }) {
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        minHeight: '200px',
        color: '#94a3b8',
      }}
    >
      <div
        style={{
          width: 40,
          height: 40,
          border: '4px solid #334155',
          borderTop: '4px solid #3b82f6',
          borderRadius: '50%',
          animation: 'spin 1s linear infinite',
          marginBottom: 16,
        }}
      />
      <div style={{ fontSize: 14 }}>{message}</div>
      <style>
        {`
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        `}
      </style>
    </div>
  )
}
