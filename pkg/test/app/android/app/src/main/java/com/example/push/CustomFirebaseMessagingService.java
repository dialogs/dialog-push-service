package com.example.push;

import android.util.Log;
import com.google.firebase.messaging.FirebaseMessagingService;

public class CustomFirebaseMessagingService extends FirebaseMessagingService {

    private static final String TAG = "MainActivity";

    @Override
    public void onCreate() {
        super.onCreate();
    }

    /**
     * Called if InstanceID token is updated. This may occur if the security of the
     * previous token had been compromised. Note that this is called when the
     * InstanceID token is initially generated so this is where you would retrieve
     * the token.
     */
    @Override
    public void onNewToken(String token) {
        Log.d(TAG, "Refreshed token: " + token);
    }
}
