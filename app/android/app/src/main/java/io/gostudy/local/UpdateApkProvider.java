package io.gostudy.local;

import android.content.ContentProvider;
import android.content.ContentValues;
import android.database.Cursor;
import android.database.MatrixCursor;
import android.net.Uri;
import android.os.ParcelFileDescriptor;
import android.provider.OpenableColumns;

import java.io.File;
import java.io.FileNotFoundException;

/**
 * 只做一件事：把已经下载好的更新包（<filesDir>/updates/GoStudy.apk）以只读方式借给系统安装器。
 *
 * 这里不引入 androidx 的 FileProvider，也不接受任何来自外部的文件路径——
 * 始终只服务那一个固定文件，安装器拿到的 URI 由代码拼出来，页面无法指定别的路径。
 */
public final class UpdateApkProvider extends ContentProvider {
    public static final String AUTHORITY = BuildConfig.APPLICATION_ID + ".updateapk";
    public static final String APK_NAME = "GoStudy.apk";
    private static final String MIME = "application/vnd.android.package-archive";

    /** 安装器要求的 MIME 类型，调用处不必再写一遍。 */
    public static String getApkMime() {
        return MIME;
    }

    @Override
    public boolean onCreate() {
        return true;
    }

    private File apkFile() {
        return new File(new File(getContext().getFilesDir(), "updates"), APK_NAME);
    }

    @Override
    public ParcelFileDescriptor openFile(Uri uri, String mode) throws FileNotFoundException {
        if (!"r".equals(mode)) {
            throw new FileNotFoundException("更新包只读");
        }
        File file = apkFile();
        if (!file.isFile()) {
            throw new FileNotFoundException("更新包不存在");
        }
        return ParcelFileDescriptor.open(file, ParcelFileDescriptor.MODE_READ_ONLY);
    }

    @Override
    public String getType(Uri uri) {
        return MIME;
    }

    @Override
    public Cursor query(Uri uri, String[] projection, String selection, String[] selectionArgs, String sortOrder) {
        File file = apkFile();
        if (!file.isFile()) {
            return null;
        }
        MatrixCursor cursor = new MatrixCursor(new String[]{OpenableColumns.DISPLAY_NAME, OpenableColumns.SIZE});
        cursor.addRow(new Object[]{APK_NAME, file.length()});
        return cursor;
    }

    @Override
    public Uri insert(Uri uri, ContentValues values) {
        throw new UnsupportedOperationException();
    }

    @Override
    public int delete(Uri uri, String selection, String[] selectionArgs) {
        throw new UnsupportedOperationException();
    }

    @Override
    public int update(Uri uri, ContentValues values, String selection, String[] selectionArgs) {
        throw new UnsupportedOperationException();
    }
}
