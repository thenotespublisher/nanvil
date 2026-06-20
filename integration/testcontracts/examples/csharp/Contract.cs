using Neo.SmartContract.Framework;

namespace NsmithExample
{
    public class Contract : SmartContract
    {
        public static string GetValue() => "nsmith-csharp-ok";
    }
}
