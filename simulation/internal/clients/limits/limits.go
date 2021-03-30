// This package is used to provide a small set of limit values
// which are not obtained from the configuration, at least at present.

package limits

const (
	// MaxBladeID is the highest blade number accepted as valid. This is
	// an arbitrary choice intended to help prevent configuration issues.
	//
	MaxBladeID = int64(10 * 1000 * 1000)

	// MaxPduID defines the larget number of Pdus that can be configured within
	// a single rack.
	//
	MaxPduID = int64(2)

	// MaxTorID defines the larget number of Tors that can be configured within
	// a single rack.
	//
	MaxTorID = int64(2)
)