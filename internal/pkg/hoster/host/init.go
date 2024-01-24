package HosterHost

// Fix time-sync issues:
// Check if chrony is installed: which chronyc
// Signal an error if chrony is not installed
// Check if chronyd is running, enable+start if it isn't
// Force-sync with the sources: chronyc -a 'burst 4/4'
// Set the time immediately: chronyc -a makestep
