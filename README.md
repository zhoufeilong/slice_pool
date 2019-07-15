# slice_pool
decrease gc cost,self pool

如果你的程序因为内存中长驻大量对象导致gc 消耗cpu严重.可以考虑使用slice缓存数据减少gc开销了。
pool.go写了一个将数据同一管理的pool类。方便项目代码的书写。
如果你的使用频率低，可以直接使用[]TestStruct,以提高效率
