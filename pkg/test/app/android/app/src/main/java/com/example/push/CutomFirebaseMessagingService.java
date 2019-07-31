package com.example.push;

import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.content.Context;
import android.content.Intent;
import android.media.RingtoneManager;
import android.net.Uri;
import android.os.Build;

import androidx.annotation.NonNull;
import androidx.core.app.NotificationCompat;
import android.util.Log;
import android.widget.TextView;

import com.google.android.gms.tasks.OnCompleteListener;
import com.google.android.gms.tasks.Task;
import com.google.firebase.iid.FirebaseInstanceId;
import com.google.firebase.iid.InstanceIdResult;
import com.google.firebase.messaging.FirebaseMessaging;
import com.google.firebase.messaging.FirebaseMessagingService;
import com.google.firebase.messaging.RemoteMessage;
import com.example.push.R;
//import activi;

//import androidx.work.OneTimeWorkRequest;
//import androidx.work.WorkManager;

public class CutomFirebaseMessagingService extends FirebaseMessagingService  {

    private static final String TAG = "MainActivity";

    @Override
    public void onCreate() {
        super.onCreate();

//        // Create channel to show notifications.
//        String channelId  = getString(R.string.default_notification_channel_id);
//        String channelName = getString(R.string.default_notification_channel_name);
//        NotificationManager notificationManager =
//                getSystemService(NotificationManager.class);
//        notificationManager.createNotificationChannel(new NotificationChannel(channelId,
//                channelName, NotificationManager.IMPORTANCE_LOW));
    }

    /**
     * Called if InstanceID token is updated. This may occur if the security of
     * the previous token had been compromised. Note that this is called when the InstanceID token
     * is initially generated so this is where you would retrieve the token.
     */
    @Override
    public void onNewToken(String token) {
        Log.d(TAG, "Refreshed token: " + token);
    }
}
