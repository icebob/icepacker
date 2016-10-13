go install github.com/icebob/icepacker

echo Pack and Unpack test1
icepacker pack r:\icepack-test\test1 r:\icepack-test\test1.pack
icepacker list r:\icepack-test\test1.pack
icepacker unpack r:\icepack-test\test1.pack r:\icepack-test\unpacked\test1

echo Pack and Unpack test2
icepacker pack -c gzip r:\icepack-test\test2 r:\icepack-test\test2.pack
icepacker list r:\icepack-test\test2.pack
icepacker unpack r:\icepack-test\test2.pack r:\icepack-test\unpacked\test2

echo Pack and Unpack test3
icepacker pack -e aes --key secretkey r:\icepack-test\test3 r:\icepack-test\test3.pack
icepacker list --key secretkey r:\icepack-test\test3.pack
icepacker unpack --key secretkey r:\icepack-test\test3.pack r:\icepack-test\unpacked\test3

echo Pack and Unpack filetest
icepacker pack -c gzip -e aes --key secretkey r:\icepack-test\filetest r:\icepack-test\filetest.pack
icepacker list --key secretkey r:\icepack-test\filetest.pack
icepacker unpack --key secretkey r:\icepack-test\filetest.pack r:\icepack-test\unpacked\filetest

echo Pack and Unpack merged file
icepacker pack r:\icepack-test\test1 r:\icepack-test\merge.pack
copy /b r:\icepack-test\bb.jpg + r:\icepack-test\merge.pack r:\icepack-test\merged.pack.jpg
icepacker unpack r:\icepack-test\merged.pack.jpg r:\icepack-test\unpacked\merged
